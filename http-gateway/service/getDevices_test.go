package service_test

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/plgd-dev/device/schema/device"
	"github.com/plgd-dev/device/schema/interfaces"
	"github.com/plgd-dev/device/test/resource/types"
	"github.com/plgd-dev/hub/grpc-gateway/pb"
	httpgwTest "github.com/plgd-dev/hub/http-gateway/test"
	"github.com/plgd-dev/hub/http-gateway/uri"
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
	"google.golang.org/grpc/credentials"
)

func makeDefaultDevice(deviceID string) *pb.Device {
	return &pb.Device{
		Types:      []string{types.DEVICE_CLOUD, device.ResourceType},
		Interfaces: []string{interfaces.OC_IF_R, interfaces.OC_IF_BASELINE},
		Id:         deviceID,
		Name:       test.TestDeviceName,
		Metadata: &pb.Device_Metadata{
			Status: &commands.ConnectionStatus{
				Value: commands.ConnectionStatus_ONLINE,
			},
		},
	}
}

func TestRequestHandlerGetDevices(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	type args struct {
		accept         string
		typeFilter     []string
		statusFilter   []pb.GetDevicesRequest_Status
		deviceIdFilter []string
	}
	tests := []struct {
		name string
		args args
		want []*pb.Device
	}{
		{
			name: "all devices",
			args: args{
				accept: uri.ApplicationProtoJsonContentType,
			},
			want: []*pb.Device{makeDefaultDevice(deviceID)},
		},
		{
			name: "offline devices",
			args: args{
				accept:       uri.ApplicationProtoJsonContentType,
				statusFilter: []pb.GetDevicesRequest_Status{pb.GetDevicesRequest_OFFLINE},
			},
		},
		{
			name: "invalid device id",
			args: args{
				accept:         uri.ApplicationProtoJsonContentType,
				deviceIdFilter: []string{"invalid"},
			},
		},
		{
			name: "single device",
			args: args{
				accept:         uri.ApplicationProtoJsonContentType,
				deviceIdFilter: []string{deviceID},
			},
			want: []*pb.Device{makeDefaultDevice(deviceID)},
		},
		{
			name: "invalid device type",
			args: args{
				accept:     uri.ApplicationProtoJsonContentType,
				typeFilter: []string{"invalid"},
			},
		},
		{
			name: "cloud device type",
			args: args{
				accept:     uri.ApplicationProtoJsonContentType,
				typeFilter: []string{types.DEVICE_CLOUD},
			},
			want: []*pb.Device{makeDefaultDevice(deviceID)},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
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

	log.Setup(log.Config{Debug: true})
	_, shutdownDevSim := test.OnboardDevSim(ctx, t, c, deviceID, config.GW_HOST, test.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	toStringSlice := func(s []pb.GetDevicesRequest_Status) []string {
		var sf []string
		for _, v := range s {
			sf = append(sf, strconv.FormatInt(int64(v), 10))
		}
		return sf
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := httpgwTest.NewRequest(http.MethodGet, uri.Devices, nil).Accept(tt.args.accept).AuthToken(token)
			rb.AddTypeFilter(tt.args.typeFilter).AddStatusFilter(toStringSlice(tt.args.statusFilter)).AddDeviceIdFilter(tt.args.deviceIdFilter)
			resp := httpgwTest.HTTPDo(t, rb.Build())
			defer func() {
				_ = resp.Body.Close()
			}()

			var devices []*pb.Device
			for {
				var dev pb.Device
				err = Unmarshal(resp.StatusCode, resp.Body, &dev)
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				assert.NotEmpty(t, dev.ProtocolIndependentId)
				devices = append(devices, &dev)
			}
			pbTest.CmpDeviceValues(t, tt.want, devices)
		})
	}
}
