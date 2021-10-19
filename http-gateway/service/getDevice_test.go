package service_test

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

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
)

func TestRequestHandler_GetDevice(t *testing.T) {
	deviceID := test.MustFindDeviceByName(test.TestDeviceName)
	type args struct {
		deviceID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []*pb.Device
	}{
		{
			name: "valid",
			args: args{
				deviceID: deviceID,
			},
			want: []*pb.Device{
				{
					Types:      []string{types.DEVICE_CLOUD, device.ResourceType},
					Interfaces: []string{interfaces.OC_IF_R, interfaces.OC_IF_BASELINE},
					Id:         deviceID,
					Name:       test.TestDeviceName,
					Metadata: &pb.Device_Metadata{
						Status: &commands.ConnectionStatus{
							Value: commands.ConnectionStatus_ONLINE,
						},
					},
				},
			},
		},
		{
			name: "notFound",
			args: args{
				deviceID: "notFound",
			},
			wantErr: true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	tearDown := test.SetUp(ctx, t)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httpgwTest.NewRequest(http.MethodGet, uri.AliasDevice+"/", nil).DeviceId(tt.args.deviceID).AuthToken(token).Build()
			trans := http.DefaultTransport.(*http.Transport).Clone()
			trans.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
			c := http.Client{
				Transport: trans,
			}
			resp, err := c.Do(request)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			devices := make([]*pb.Device, 0, 1)
			for {
				var dev pb.Device
				err = Unmarshal(resp.StatusCode, resp.Body, &dev)
				if err == io.EOF {
					break
				}
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.NotEmpty(t, dev.ProtocolIndependentId)
				dev.ProtocolIndependentId = ""
				dev.Metadata.Status.ValidUntil = 0
				devices = append(devices, &dev)
			}
			test.CheckProtobufs(t, tt.want, devices, test.RequireToCheckFunc(require.Equal))
		})
	}
}
