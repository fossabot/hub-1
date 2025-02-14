package service_test

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"testing"

	"github.com/plgd-dev/device/schema/collection"
	"github.com/plgd-dev/device/schema/configuration"
	"github.com/plgd-dev/device/schema/device"
	"github.com/plgd-dev/device/schema/interfaces"
	"github.com/plgd-dev/device/schema/platform"
	"github.com/plgd-dev/device/test/resource/types"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	httpgwTest "github.com/plgd-dev/hub/http-gateway/test"
	"github.com/plgd-dev/hub/http-gateway/uri"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/plgd-dev/hub/test"
	"github.com/plgd-dev/hub/test/config"
	oauthTest "github.com/plgd-dev/hub/test/oauth-server/test"
	pbTest "github.com/plgd-dev/hub/test/pb"
	"github.com/plgd-dev/hub/test/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func makePlatformResourceChanged(t *testing.T, deviceID string) *events.ResourceChanged {
	return pbTest.MakeResourceChanged(t, deviceID, platform.ResourceURI,
		map[string]interface{}{
			"mnmn": "ocfcloud.com",
			//"pi":   "d9b71824-78f7-4f26-540b-d86eab696937",
			"if": []interface{}{interfaces.OC_IF_R, interfaces.OC_IF_BASELINE},
			"rt": []interface{}{platform.ResourceType},
		},
	)
}

func makeCloudDeviceResourceChanged(t *testing.T, deviceID string) *events.ResourceChanged {
	return pbTest.MakeResourceChanged(t, deviceID, device.ResourceURI,
		map[string]interface{}{
			"n":   test.TestDeviceName,
			"di":  deviceID,
			"dmv": "ocf.res.1.3.0",
			"icv": "ocf.2.0.5",
			// "piid": "1dcb14bd-5167-4122-6c2f-71741543fdc3",
			"if": []interface{}{interfaces.OC_IF_R, interfaces.OC_IF_BASELINE},
			"rt": []interface{}{types.DEVICE_CLOUD, device.ResourceType},
		},
	)
}

func TestRequestHandlerGetDeviceResources(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	type args struct {
		deviceID   string
		typeFilter []string
		accept     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []*pb.Resource
	}{
		{
			name: "get resource of " + deviceID,
			args: args{
				deviceID: deviceID,
				accept:   uri.ApplicationProtoJsonContentType,
			},
			want: []*pb.Resource{
				{
					Types: []string{types.CORE_LIGHT},
					Data: pbTest.MakeResourceChanged(t, deviceID, test.TestResourceLightInstanceHref("1"),
						map[string]interface{}{
							"state": false,
							"power": uint64(0),
							"name":  "Light",
							"if":    []interface{}{interfaces.OC_IF_RW, interfaces.OC_IF_BASELINE},
							"rt":    []interface{}{types.CORE_LIGHT},
						},
					),
				},
				{
					Types: []string{collection.ResourceType},
					Data: pbTest.MakeResourceChanged(t, deviceID, test.TestResourceSwitchesHref,
						map[string]interface{}{
							"links":                     []interface{}{},
							"if":                        []interface{}{interfaces.OC_IF_LL, interfaces.OC_IF_CREATE, interfaces.OC_IF_B, interfaces.OC_IF_BASELINE},
							"rt":                        []interface{}{collection.ResourceType},
							"rts":                       []interface{}{types.BINARY_SWITCH},
							"rts-m":                     []interface{}{types.BINARY_SWITCH},
							"x.org.openconnectivity.bl": uint64(94),
						},
					),
				},
				{
					Types: []string{configuration.ResourceType},
					Data: pbTest.MakeResourceChanged(t, deviceID, configuration.ResourceURI,
						map[string]interface{}{
							"n":  test.TestDeviceName,
							"if": []interface{}{interfaces.OC_IF_RW, interfaces.OC_IF_BASELINE},
							"rt": []interface{}{configuration.ResourceType},
						},
					),
				},
				{
					Types: []string{platform.ResourceType},
					Data:  makePlatformResourceChanged(t, deviceID),
				},
				{
					Types: []string{types.DEVICE_CLOUD, device.ResourceType},
					Data:  makeCloudDeviceResourceChanged(t, deviceID),
				},
			},
		},
		{
			name: "get oic.wk.d and oic.wk.p of " + deviceID,
			args: args{
				deviceID:   deviceID,
				typeFilter: []string{device.ResourceType, platform.ResourceType},
				accept:     uri.ApplicationProtoJsonContentType,
			},
			want: []*pb.Resource{
				{
					Types: []string{platform.ResourceType},
					Data:  makePlatformResourceChanged(t, deviceID),
				},
				{
					Types: []string{types.DEVICE_CLOUD, device.ResourceType},
					Data:  makeCloudDeviceResourceChanged(t, deviceID),
				},
			},
		},
		{
			name: "not found",
			args: args{
				deviceID: "notFound",
				accept:   uri.ApplicationProtoJsonContentType,
			},
			wantErr: true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.TEST_TIMEOUT)
	defer cancel()

	tearDown := service.SetUp(ctx, t)
	defer tearDown()

	shutdownHttp := httpgwTest.SetUp(t)
	defer shutdownHttp()

	token := oauthTest.GetDefaultServiceToken(t)
	ctx = kitNetGrpc.CtxWithToken(ctx, token)

	conn, err := grpc.Dial(config.GRPC_HOST, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: test.GetRootCertificatePool(t),
	})))
	require.NoError(t, err)
	c := pb.NewGrpcGatewayClient(conn)

	_, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := httpgwTest.NewRequest(http.MethodGet, uri.AliasDeviceResources, nil).Accept(tt.args.accept).AuthToken(token)
			rb.DeviceId(tt.args.deviceID).AddTypeFilter(tt.args.typeFilter)
			resp := httpgwTest.HTTPDo(t, rb.Build())
			defer func() {
				_ = resp.Body.Close()
			}()

			values := make([]*pb.Resource, 0, 1)
			for {
				var value pb.Resource
				err = Unmarshal(resp.StatusCode, resp.Body, &value)
				if err == io.EOF {
					break
				}
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				values = append(values, &value)
			}
			pbTest.CmpResourceValues(t, tt.want, values)
		})
	}
}
