package service

import (
	"fmt"

	pbAS "github.com/go-ocf/cloud/authorization/pb"
	"github.com/go-ocf/cloud/coap-gateway/coapconv"
	pbCQRS "github.com/go-ocf/cloud/resource-aggregate/pb"
	"github.com/go-ocf/go-coap/v2/message"
	coapCodes "github.com/go-ocf/go-coap/v2/message/codes"
	"github.com/go-ocf/go-coap/v2/mux"
	"github.com/go-ocf/kit/codec/cbor"
	"github.com/go-ocf/kit/log"
	"github.com/go-ocf/kit/net/coap"
	kitNetGrpc "github.com/go-ocf/kit/net/grpc"
	"google.golang.org/grpc/status"
)

type CoapSignInReq struct {
	DeviceID    string `json:"di"`
	UserID      string `json:"uid"`
	AccessToken string `json:"accesstoken"`
	Login       bool   `json:"login"`
}

type CoapSignInResp struct {
	ExpiresIn int64 `json:"expiresin"`
}

// https://github.com/openconnectivityfoundation/security/blob/master/swagger2.0/oic.sec.session.swagger.json
func signInPostHandler(s mux.ResponseWriter, req *mux.Message, client *Client, signIn CoapSignInReq) {
	resp, err := client.server.asClient.SignIn(req.Context, &pbAS.SignInRequest{
		DeviceId:    signIn.DeviceID,
		UserId:      signIn.UserID,
		AccessToken: signIn.AccessToken,
	})
	if err != nil {
		client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign in: %v", err), coapconv.GrpcCode2CoapCode(status.Convert(err).Code(), coapCodes.POST), req.Token)
		return
	}

	coapResp := CoapSignInResp{
		ExpiresIn: resp.ExpiresIn,
	}

	accept := coap.GetAccept(req.Options)
	encode, err := coap.GetEncoder(accept)
	if err != nil {
		client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign in: %v", err), coapCodes.InternalServerError, req.Token)
		return
	}
	out, err := encode(coapResp)
	if err != nil {
		client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign in: %v", err), coapCodes.InternalServerError, req.Token)
		return
	}

	authCtx := authCtx{
		AuthorizationContext: pbCQRS.AuthorizationContext{
			DeviceId: signIn.DeviceID,
		},
		UserID:      signIn.UserID,
		AccessToken: signIn.AccessToken,
	}
	serviceToken, err := client.server.oauthMgr.GetToken(req.Context)
	if err != nil {
		client.logAndWriteErrorResponse(fmt.Errorf("cannot get service token: %v", err), coapCodes.InternalServerError, req.Token)
		client.Close()
		return
	}
	req.Context = kitNetGrpc.CtxWithUserID(kitNetGrpc.CtxWithToken(req.Context, serviceToken.AccessToken), authCtx.UserID)
	err = client.UpdateCloudDeviceStatus(req.Context, signIn.DeviceID, authCtx.AuthorizationContext, true)
	if err != nil {
		// Events from resources of device will be comes but device is offline. To recover cloud state, client need to reconnect to cloud.
		client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign in: cannot update cloud device status: %v", err), coapCodes.InternalServerError, req.Token)
		client.Close()
		return
	}

	oldAuthCtx := client.replaceAuthorizationContext(authCtx)
	newDevice := false

	switch {
	case oldAuthCtx.GetDeviceId() == "":
		newDevice = true
	case oldAuthCtx.GetDeviceId() != signIn.DeviceID:
		wait, err := client.server.userDevicesSubscription.Cancel(oldAuthCtx.UserID, oldAuthCtx.GetDeviceId())
		if err != nil {
			client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign in: %v", err), coapCodes.InternalServerError, req.Token)
			client.Close()
			return
		}
		wait()
		newDevice = true
	}

	if newDevice {
		sub := &deviceSubscriptionHandlers{
			Client:   client,
			deviceID: signIn.DeviceID,
		}
		_, err := client.server.userDevicesSubscription.Create(req.Context, signIn.UserID, signIn.DeviceID, sub)
		if err != nil {
			client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign in: %v", err), coapCodes.InternalServerError, req.Token)
			client.Close()
			return
		}
	}
	client.sendResponse(coapCodes.Changed, req.Token, accept, out)
}

func signOutPostHandler(s mux.ResponseWriter, req *mux.Message, client *Client, signOut CoapSignInReq) {
	_, err := client.server.asClient.SignOut(req.Context, &pbAS.SignOutRequest{
		DeviceId:    signOut.DeviceID,
		UserId:      signOut.UserID,
		AccessToken: signOut.AccessToken,
	})
	if err != nil {
		client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign out: %w", err), coapconv.GrpcCode2CoapCode(status.Convert(err).Code(), coapCodes.POST), req.Token)
		client.Close()
		return
	}

	oldAuthCtx := client.replaceAuthorizationContext(authCtx{})
	if oldAuthCtx.DeviceId != "" {
		serviceToken, err := client.server.oauthMgr.GetToken(req.Context)
		if err != nil {
			client.logAndWriteErrorResponse(fmt.Errorf("cannot get service token: %v", err), coapCodes.InternalServerError, req.Token)
			client.Close()
			return
		}
		req.Context = kitNetGrpc.CtxWithUserID(kitNetGrpc.CtxWithToken(req.Context, serviceToken.AccessToken), oldAuthCtx.UserID)
		err = client.UpdateCloudDeviceStatus(req.Context, oldAuthCtx.DeviceId, oldAuthCtx.AuthorizationContext, false)
		if err != nil {
			// Device will be still reported as online and it can fix his state by next calls online, offline commands.
			log.Errorf("DeviceId %v: cannot handle sign out: cannot update cloud device status: %v", oldAuthCtx.GetDeviceId(), err)
			return
		}

		wait, err := client.server.userDevicesSubscription.Cancel(oldAuthCtx.UserID, oldAuthCtx.GetDeviceId())
		if err != nil {
			client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign out: %w", err), coapCodes.InternalServerError, req.Token)
			client.Close()
			return
		}
		wait()
	}

	client.sendResponse(coapCodes.Changed, req.Token, message.AppOcfCbor, []byte{0xA0}) // empty object
}

// Sign-in
// https://github.com/openconnectivityfoundation/security/blob/master/swagger2.0/oic.sec.session.swagger.json
func signInHandler(s mux.ResponseWriter, req *mux.Message, client *Client) {
	switch req.Code {
	case coapCodes.POST:
		var signIn CoapSignInReq
		err := cbor.ReadFrom(req.Body, &signIn)
		if err != nil {
			client.logAndWriteErrorResponse(fmt.Errorf("cannot handle sign in: %v", err), coapCodes.BadRequest, req.Token)
			return
		}
		switch signIn.Login {
		case true:
			signInPostHandler(s, req, client, signIn)
		default:
			signOutPostHandler(s, req, client, signIn)
		}
	default:
		client.logAndWriteErrorResponse(fmt.Errorf("Forbidden request from %v", client.remoteAddrString()), coapCodes.Forbidden, req.Token)
	}
}
