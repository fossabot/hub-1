package service_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	coapCodes "github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/hub/v2/coap-gateway/service"
	coapgwTest "github.com/plgd-dev/hub/v2/coap-gateway/test"
	"github.com/plgd-dev/hub/v2/coap-gateway/uri"
	"github.com/plgd-dev/hub/v2/grpc-gateway/pb"
	kitNetGrpc "github.com/plgd-dev/hub/v2/pkg/net/grpc"
	test "github.com/plgd-dev/hub/v2/test"
	testCfg "github.com/plgd-dev/hub/v2/test/config"
	oauthTest "github.com/plgd-dev/hub/v2/test/oauth-server/test"
	oauthUri "github.com/plgd-dev/hub/v2/test/oauth-server/uri"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TestCoapSignInResponse struct {
	ExpiresIn uint64 `json:"-"`
}

func TestSignInPostHandler(t *testing.T) {
	shutdown := setUp(t)
	defer shutdown()

	tbl := []testEl{
		{"BadRequest (invalid request)", input{coapCodes.POST, `{"login": true}`, nil}, output{coapCodes.BadRequest, `invalid device id`, nil}, true},
		{"Unauthorized (invalid userID)", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid": "0", "accesstoken":"%%ACCESS_TOKEN%%", "login": true }`, nil}, output{coapCodes.Unauthorized, `doesn't match userID`, nil}, true},
		{"BadRequest (missing access token)", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid": "0", "login": true }`, nil}, output{coapCodes.BadRequest, `invalid access token`, nil}, true},
		{"BadRequest (invalid access token)", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "accesstoken": 123, "login": true}`, nil}, output{coapCodes.BadRequest, `cannot handle sign in: cannot decode body: cbor`, nil}, true},
		{"Changed1", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid":"%%USER_ID%%", "accesstoken":"%%ACCESS_TOKEN%%", "login": true }`, nil}, output{coapCodes.Changed, TestCoapSignInResponse{}, nil}, false},
	}

	for _, test := range tbl {
		tf := func(t *testing.T) {
			co := testCoapDial(t, testCfg.GW_HOST, "", true, time.Now().Add(time.Minute))
			if co == nil {
				return
			}
			defer func() {
				_ = co.Close()
			}()
			signUpResp := testSignUp(t, CertIdentity, co)

			payload := test.in.payload.(string)
			payload = strings.Replace(payload, "%%USER_ID%%", signUpResp.UserID, -1)
			payload = strings.Replace(payload, "%%ACCESS_TOKEN%%", signUpResp.AccessToken, -1)
			test.in.payload = payload

			testPostHandler(t, uri.SignIn, test, co)
		}
		t.Run(test.name, tf)
	}
}

func TestSignInDeviceSubscriptionHandler(t *testing.T) {
	shutdown := setUp(t)
	defer shutdown()

	ctx := kitNetGrpc.CtxWithToken(context.Background(), oauthTest.GetDefaultAccessToken(t))
	conn, err := grpc.Dial(testCfg.GRPC_HOST, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: test.GetRootCertificatePool(t),
	})))
	require.NoError(t, err)
	c := pb.NewGrpcGatewayClient(conn)
	defer func() {
		_ = conn.Close()
	}()

	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	co := testCoapDial(t, testCfg.GW_HOST, "", true, time.Now().Add(time.Minute))
	if co == nil {
		return
	}
	defer func() {
		_ = co.Close()
	}()

	co.AddOnClose(func() {
		cancel()
	})

	signUpResp := testSignUp(t, CertIdentity, co)
	testSignIn(t, CertIdentity, signUpResp, co)
	_, err = c.DeleteDevices(ctx, &pb.DeleteDevicesRequest{
		DeviceIdFilter: []string{CertIdentity},
	})
	require.NoError(t, err)

	<-cancelCtx.Done()
	require.True(t, errors.Is(cancelCtx.Err(), context.Canceled))

	co1 := testCoapDial(t, testCfg.GW_HOST, "", true, time.Now().Add(time.Minute))
	resp, err := doSignIn(t, CertIdentity, signUpResp, co1)
	if err != nil {
		require.Contains(t, err.Error(), "context canceled")
		return
	}
	require.Equal(t, coapCodes.Unauthorized, resp.Code())
	_ = co1.Close()
}

func TestSignOutPostHandler(t *testing.T) {
	shutdown := setUp(t)
	defer shutdown()

	co := testCoapDial(t, testCfg.GW_HOST, "", true, time.Now().Add(time.Minute))
	if co == nil {
		return
	}
	defer func() {
		_ = co.Close()
	}()

	signUpResp := testSignUp(t, CertIdentity, co)
	testSignIn(t, CertIdentity, signUpResp, co)

	tbl := []testEl{
		{"Changed (uid from ctx)", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "accesstoken":"` + signUpResp.AccessToken + `", "login": false }`, nil}, output{coapCodes.Changed, TestCoapSignInResponse{}, nil}, false},
		{"Changed1", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid":"` + signUpResp.UserID + `", "accesstoken":"` + signUpResp.AccessToken + `", "login": false }`, nil}, output{coapCodes.Changed, TestCoapSignInResponse{}, nil}, false},
	}

	for _, test := range tbl {
		tf := func(t *testing.T) {
			testPostHandler(t, uri.SignIn, test, co)
		}
		t.Run(test.name, tf)
	}
}

func TestSignInWithMTLSAndDeviceIdClaim(t *testing.T) {
	coapgwCfg := coapgwTest.MakeConfig(t)
	coapgwCfg.APIs.COAP.TLS.Enabled = true
	coapgwCfg.APIs.COAP.TLS.Embedded.ClientCertificateRequired = true
	coapgwCfg.APIs.COAP.Authorization.DeviceIDClaim = oauthUri.DeviceIDClaimKey
	shutdown := setUp(t, coapgwCfg)
	defer shutdown()

	signUp := func(deviceID string) service.CoapSignUpResponse {
		co := testCoapDial(t, testCfg.GW_HOST, deviceID, true, time.Now().Add(time.Minute))
		require.NotEmpty(t, co)
		signUpResp := testSignUp(t, deviceID, co)
		_ = co.Close()
		return signUpResp
	}

	signUpResp := signUp(CertIdentity)
	anotherDeviceID := uuid.New().String()

	check := func(deviceID string, req testEl) {
		co := testCoapDial(t, testCfg.GW_HOST, deviceID, true, time.Now().Add(time.Minute))
		require.NotEmpty(t, co)
		testPostHandler(t, uri.SignIn, req, co)
		_ = co.Close()
	}

	tokenWithoutDeviceID := oauthTest.GetDefaultAccessToken(t)

	req := testEl{"OK", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid":"` + signUpResp.UserID + `", "accesstoken":"` + signUpResp.AccessToken + `", "login": true }`, nil}, output{coapCodes.Changed, TestCoapSignInResponse{}, nil}, false}
	check(CertIdentity, req)

	req = testEl{"mtls deviceID != JWT deviceID", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid":"` + signUpResp.UserID + `", "accesstoken":"` + signUpResp.AccessToken + `", "login": true }`, nil}, output{coapCodes.Unauthorized, `cannot handle sign in: access token issued to the device`, nil}, true}
	check(anotherDeviceID, req)

	req = testEl{"JWT deviceID is not set", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid":"` + signUpResp.UserID + `", "accesstoken":"` + tokenWithoutDeviceID + `", "login": true }`, nil}, output{coapCodes.Unauthorized, `cannot handle sign in: access token doesn't contain the required device id claim`, nil}, true}
	check(CertIdentity, req)
}

func TestCertificateExpiration(t *testing.T) {
	coapgwCfg := coapgwTest.MakeConfig(t)
	coapgwCfg.APIs.COAP.TLS.Enabled = true
	coapgwCfg.APIs.COAP.OwnerCacheExpiration = time.Second
	coapgwCfg.APIs.COAP.TLS.DisconnectOnExpiredCertificate = true
	coapgwCfg.APIs.COAP.TLS.Embedded.ClientCertificateRequired = true
	coapgwCfg.APIs.COAP.Authorization.DeviceIDClaim = oauthUri.DeviceIDClaimKey

	shutdown := setUp(t, coapgwCfg)
	defer shutdown()

	signUp := func(deviceID string) service.CoapSignUpResponse {
		co := testCoapDial(t, testCfg.GW_HOST, deviceID, true, time.Now().Add(time.Minute))
		require.NotEmpty(t, co)
		signUpResp := testSignUp(t, deviceID, co)
		_ = co.Close()
		return signUpResp
	}

	signUpResp := signUp(CertIdentity)

	duration := time.Second * 4

	req := testEl{"OK", input{coapCodes.POST, `{"di": "` + CertIdentity + `", "uid":"` + signUpResp.UserID + `", "accesstoken":"` + signUpResp.AccessToken + `", "login": true }`, nil}, output{coapCodes.Changed, TestCoapSignInResponse{}, nil}, false}
	co := testCoapDial(t, testCfg.GW_HOST, CertIdentity, true, time.Now().Add(duration))
	require.NotEmpty(t, co)
	defer func() {
		_ = co.Close()
	}()
	testPostHandler(t, uri.SignIn, req, co)

	select {
	case <-co.Done():
		// connection was closed by certificate expiration
		return
	case <-time.After(2 * duration):
		require.NoError(t, fmt.Errorf("timeout"))
	}
}
