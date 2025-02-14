package service_test

import (
	"context"
	"crypto/tls"
	"io"
	"testing"

	"github.com/plgd-dev/device/schema/device"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/resource-aggregate/commands"
	"github.com/plgd-dev/hub/resource-aggregate/events"
	"github.com/plgd-dev/hub/test"
	testCfg "github.com/plgd-dev/hub/test/config"
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
		req *pb.GetDevicesMetadataRequest
	}
	tests := []struct {
		name    string
		args    args
		want    []*events.DeviceMetadataUpdated
		wantErr bool
	}{
		{
			name: "all",
			args: args{
				req: &pb.GetDevicesMetadataRequest{},
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
			name: "filter one device",
			args: args{
				req: &pb.GetDevicesMetadataRequest{
					DeviceIdFilter: []string{deviceID},
				},
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
				req: &pb.GetDevicesMetadataRequest{
					TypeFilter: []string{device.ResourceType},
				},
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
				req: &pb.GetDevicesMetadataRequest{
					DeviceIdFilter: []string{"abc"},
				},
			},
			wantErr: true,
		},
		{
			name: "unknown type",
			args: args{
				req: &pb.GetDevicesMetadataRequest{
					TypeFilter: []string{"unknown"},
				},
			},
			wantErr: true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), testCfg.TEST_TIMEOUT)
	defer cancel()

	tearDown := service.SetUp(ctx, t)
	defer tearDown()
	ctx = kitNetGrpc.CtxWithToken(ctx, oauthTest.GetDefaultServiceToken(t))

	conn, err := grpc.Dial(testCfg.GRPC_HOST, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: test.GetRootCertificatePool(t),
	})))
	require.NoError(t, err)
	c := pb.NewGrpcGatewayClient(conn)

	_, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, testCfg.GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := c.GetDevicesMetadata(ctx, tt.args.req)
			require.NoError(t, err)
			var values []*events.DeviceMetadataUpdated
			for {
				value, err := client.Recv()
				if err == io.EOF {
					break
				}
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				values = append(values, value)
			}
			pbTest.CmpDeviceMetadataUpdatedSlice(t, tt.want, values)
		})
	}
}
