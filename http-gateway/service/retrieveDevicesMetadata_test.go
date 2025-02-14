package service_test

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"testing"

	"github.com/plgd-dev/device/schema/device"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	httpgwTest "github.com/plgd-dev/hub/http-gateway/test"
	"github.com/plgd-dev/hub/http-gateway/uri"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/plgd-dev/hub/test"
	"github.com/plgd-dev/hub/test/config"
	oauthService "github.com/plgd-dev/hub/test/oauth-server/service"
	oauthTest "github.com/plgd-dev/hub/test/oauth-server/test"
	pbTest "github.com/plgd-dev/hub/test/pb"
	"github.com/plgd-dev/hub/test/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestRequestHandlerGetDevicesMetadata(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	type args struct {
		typeFilter     []string
		deviceIdFilter []string
	}
	tests := []struct {
		name string
		args args
		want []*events.DeviceMetadataUpdated
	}{
		{
			name: "all",
			args: args{},
			want: []*events.DeviceMetadataUpdated{
				{
					DeviceId: deviceID,
					Status: &commands.ConnectionStatus{
						Value: commands.ConnectionStatus_ONLINE,
					},
					AuditContext: commands.NewAuditContext(oauthService.DeviceUserID, ""),
				},
			},
		},
		{
			name: "filter one device",
			args: args{
				deviceIdFilter: []string{deviceID},
			},
			want: []*events.DeviceMetadataUpdated{
				{
					DeviceId: deviceID,
					Status: &commands.ConnectionStatus{
						Value: commands.ConnectionStatus_ONLINE,
					},
					AuditContext: commands.NewAuditContext(oauthService.DeviceUserID, ""),
				},
			},
		},
		{
			name: "filter one device by type",
			args: args{
				typeFilter: []string{device.ResourceType},
			},
			want: []*events.DeviceMetadataUpdated{
				{
					DeviceId: deviceID,
					Status: &commands.ConnectionStatus{
						Value: commands.ConnectionStatus_ONLINE,
					},
					AuditContext: commands.NewAuditContext(oauthService.DeviceUserID, ""),
				},
			},
		},
		{
			name: "invalid deviceID",
			args: args{
				deviceIdFilter: []string{"abc"},
			},
		},
		{
			name: "unknown type",
			args: args{
				typeFilter: []string{"unknown"},
			},
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
	defer func() {
		_ = conn.Close()
	}()

	_, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := httpgwTest.NewRequest(http.MethodGet, uri.DevicesMetadata, nil).AuthToken(token)
			rb.AddTypeFilter(tt.args.typeFilter).AddDeviceIdFilter(tt.args.deviceIdFilter)
			resp := httpgwTest.HTTPDo(t, rb.Build())
			defer func() {
				_ = resp.Body.Close()
			}()

			var values []*events.DeviceMetadataUpdated
			for {
				var value events.DeviceMetadataUpdated
				err = Unmarshal(resp.StatusCode, resp.Body, &value)
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				values = append(values, &value)
			}
			pbTest.CmpDeviceMetadataUpdatedSlice(t, tt.want, values)
		})
	}
}
