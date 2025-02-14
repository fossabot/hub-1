package service_test

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/plgd-dev/device/schema/device"
	"github.com/plgd-dev/device/schema/interfaces"
	"github.com/plgd-dev/device/test/resource/types"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	"github.com/plgd-dev/hub/pkg/log"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/test"
	"github.com/plgd-dev/hub/test/config"
	oauthTest "github.com/plgd-dev/hub/test/oauth-server/test"
	pbTest "github.com/plgd-dev/hub/test/pb"
	"github.com/plgd-dev/hub/test/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func makeCreateResourceRequest(t *testing.T, deviceID, href string, data map[string]interface{}, ttl int64) *pb.CreateResourceRequest {
	return &pb.CreateResourceRequest{
		ResourceId: commands.NewResourceID(deviceID, href),
		Content: &pb.Content{
			ContentType: message.AppOcfCbor.String(),
			Data:        test.EncodeToCbor(t, data),
		},
		TimeToLive: ttl,
	}
}

func TestRequestHandlerCreateResource(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	type args struct {
		href string
		data map[string]interface{}
		ttl  int64
	}
	tests := []struct {
		name        string
		args        args
		wantData    map[string]interface{}
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "invalid Href",
			args: args{
				href: "/unknown",
				data: map[string]interface{}{
					"power": 1,
				},
			},
			wantErr:     true,
			wantErrCode: codes.NotFound,
		},
		{
			name: "/oic/d - PermissionDenied",
			args: args{
				href: device.ResourceURI,
				data: map[string]interface{}{
					"power": 1,
				},
			},
			wantErr:     true,
			wantErrCode: codes.PermissionDenied,
		},
		{
			name: "invalid timeToLive",
			args: args{
				href: device.ResourceURI,
				data: map[string]interface{}{
					"power": 1,
				},
				ttl: int64(99 * time.Millisecond),
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "missing if",
			args: args{
				href: test.TestResourceSwitchesHref,
				data: map[string]interface{}{
					"rt": []interface{}{types.BINARY_SWITCH},
					"rep": map[string]interface{}{
						"value": false,
					},
				},
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "missing rt",
			args: args{
				href: test.TestResourceSwitchesHref,
				data: map[string]interface{}{
					"if": []interface{}{interfaces.OC_IF_A, interfaces.OC_IF_BASELINE},
					"rep": map[string]interface{}{
						"value": false,
					},
				},
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "missing rep",
			args: args{
				href: test.TestResourceSwitchesHref,
				data: map[string]interface{}{
					"if": []interface{}{interfaces.OC_IF_A, interfaces.OC_IF_BASELINE},
					"rt": []interface{}{types.BINARY_SWITCH},
				},
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "create /switches/1",
			args: args{
				href: test.TestResourceSwitchesHref,
				data: test.MakeSwitchResourceDefaultData(),
			},
			wantData: pbTest.MakeCreateLightResourceResponseData("1"),
		},
		{
			name: "create /switches/2",
			args: args{
				href: test.TestResourceSwitchesHref,
				data: test.MakeSwitchResourceDefaultData(),
			},
			wantData: pbTest.MakeCreateLightResourceResponseData("2"),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.TEST_TIMEOUT)
	defer cancel()

	tearDown := service.SetUp(ctx, t)
	defer tearDown()
	log.Setup(log.Config{Debug: true})
	ctx = kitNetGrpc.CtxWithToken(ctx, oauthTest.GetDefaultServiceToken(t))

	conn, err := grpc.Dial(config.GRPC_HOST, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: test.GetRootCertificatePool(t),
	})))
	require.NoError(t, err)
	c := pb.NewGrpcGatewayClient(conn)

	_, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := makeCreateResourceRequest(t, deviceID, tt.args.href, tt.args.data, tt.args.ttl)
			got, err := c.CreateResource(ctx, req)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.wantErrCode.String(), status.Convert(err).Code().String())
				return
			}
			require.NoError(t, err)
			resp := pbTest.MakeResourceCreated(t, deviceID, tt.args.href, tt.wantData)
			pbTest.CmpResourceCreated(t, resp, got.GetData())
		})
	}
}
