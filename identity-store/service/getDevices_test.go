package service

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/plgd-dev/hub/identity-store/pb"
	kitNetGrpc "github.com/plgd-dev/hub/pkg/net/grpc"
	"github.com/plgd-dev/hub/test/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestUserDevicesList(t *testing.T) {
	jwtWithSubTestUserID := config.CreateJwtToken(t, jwt.MapClaims{
		"sub": testUserID,
	})
	srv := newMockRetrieveResources(kitNetGrpc.CtxWithIncomingToken(context.Background(), jwtWithSubTestUserID))
	s, shutdown := newTestService(t)
	defer shutdown()
	defer func() {
		err := s.cleanUp()
		require.NoError(t, err)
	}()
	persistDevice(t, s.service.persistence, newTestDevice())
	err := s.service.GetDevices(newGetDevicesRequest(), srv)
	assert.NoError(t, err)
	r := map[string]*pb.Device{
		testDeviceID: {
			DeviceId: testDeviceID,
		},
	}
	assert.Equal(t, r, srv.resourceValues)
}

func TestListingMoreDevices(t *testing.T) {
	jwtWithSubTestUserID := config.CreateJwtToken(t, jwt.MapClaims{
		"sub": testUserID,
	})
	srv := newMockRetrieveResources(kitNetGrpc.CtxWithIncomingToken(context.Background(), jwtWithSubTestUserID))
	s, shutdown := newTestService(t)
	defer shutdown()
	defer func() {
		err := s.cleanUp()
		require.NoError(t, err)
	}()
	persistDevice(t, s.service.persistence, newTestDevice())
	d := newTestDevice()
	d.DeviceID = "anotherDeviceID"
	persistDevice(t, s.service.persistence, d)

	err := s.service.GetDevices(newGetDevicesRequest(), srv)
	assert := assert.New(t)
	assert.NoError(err)
	r := map[string]*pb.Device{
		testDeviceID: {
			DeviceId: testDeviceID,
		},
		d.DeviceID: {
			DeviceId: d.DeviceID,
		},
	}
	assert.Equal(r, srv.resourceValues)
}

func newGetDevicesRequest() *pb.GetDevicesRequest {
	return &pb.GetDevicesRequest{}
}

type mockGeDevicesServer struct {
	resourceValues map[string]*pb.Device
	ctx            context.Context
	grpc.ServerStream
}

func newMockRetrieveResources(ctx context.Context) *mockGeDevicesServer {
	return &mockGeDevicesServer{
		ctx: ctx,
	}
}

func (d *mockGeDevicesServer) Send(r *pb.Device) error {
	if d.resourceValues == nil {
		d.resourceValues = make(map[string]*pb.Device)
	}
	d.resourceValues[r.DeviceId] = r
	return nil
}

func (d *mockGeDevicesServer) Context() context.Context {
	return d.ctx
}
