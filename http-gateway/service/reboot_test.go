package service_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"
	"testing"

	authTest "github.com/go-ocf/cloud/authorization/provider"
	"github.com/go-ocf/cloud/grpc-gateway/pb"
	"github.com/go-ocf/cloud/http-gateway/test"
	"github.com/go-ocf/cloud/http-gateway/uri"
	cloudTest "github.com/go-ocf/cloud/test"
	testCfg "github.com/go-ocf/cloud/test/config"
	kitNetGrpc "github.com/go-ocf/kit/net/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestDeviceReboot(t *testing.T) {
	deviceID := cloudTest.MustFindDeviceByName(cloudTest.TestDeviceName)
	ctx, cancel := context.WithTimeout(context.Background(), test.TestTimeout)
	defer cancel()
	ctx = kitNetGrpc.CtxWithToken(ctx, authTest.UserToken)
	tearDown := cloudTest.SetUp(ctx, t)
	defer tearDown()

	conn, err := grpc.Dial(testCfg.GRPC_HOST, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs: cloudTest.GetRootCertificatePool(t),
	})))
	require.NoError(t, err)
	c := pb.NewGrpcGatewayClient(conn)
	defer conn.Close()
	shutdownDevSim := cloudTest.OnboardDevSim(ctx, t, c, deviceID, testCfg.GW_HOST, cloudTest.GetAllBackendResourceLinks())
	defer shutdownDevSim()

	webTearDown := test.SetUp(t)
	defer webTearDown()

	req := test.NewRequest("POST", uri.DeviceReboot, strings.NewReader("")).DeviceId(deviceID).AuthToken(authTest.UserToken).Build()
	res := test.HTTPDo(t, req)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}
