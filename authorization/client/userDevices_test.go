package client

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-ocf/cloud/authorization/pb"
	"github.com/go-ocf/cloud/authorization/service"
	authService "github.com/go-ocf/cloud/authorization/test"
	"github.com/go-ocf/kit/security/certManager"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type testTrigger struct {
	sync.Mutex
	addedDevices   map[string]map[string]bool
	removedDevices map[string]map[string]bool
	allDevices     map[string]map[string]bool
}

func (t *testTrigger) Clone() *testTrigger {
	t.Lock()
	defer t.Unlock()
	a := newTestTrigger()
	for userID, addedDevices := range t.addedDevices {
		a.Trigger(nil, userID, addedDevices, nil, nil)
	}
	for userID, removedDevices := range t.removedDevices {
		a.Trigger(nil, userID, nil, removedDevices, nil)
	}
	for userID, allDevices := range t.allDevices {
		a.Trigger(nil, userID, nil, nil, allDevices)
	}

	return a
}

func newTestTrigger() *testTrigger {
	return &testTrigger{}
}

func (t *testTrigger) Trigger(ctx context.Context, userID string, addedDevices, removedDevices, allDevices map[string]bool) {
	t.Lock()
	defer t.Unlock()
	if len(addedDevices) > 0 {
		if t.addedDevices == nil {
			t.addedDevices = make(map[string]map[string]bool)
		}
		devices, ok := t.addedDevices[userID]
		if !ok {
			devices = make(map[string]bool)
			t.addedDevices[userID] = devices
		}
		for deviceID := range addedDevices {
			devices[deviceID] = true
		}
	}
	if len(removedDevices) > 0 {
		if t.removedDevices == nil {
			t.removedDevices = make(map[string]map[string]bool)
		}
		devices, ok := t.removedDevices[userID]
		if !ok {
			devices = make(map[string]bool)
			t.removedDevices[userID] = devices
		}
		for deviceID := range removedDevices {
			devices[deviceID] = true
		}
	}
	if len(allDevices) == 0 {
		t.allDevices = nil
		return
	}
	if t.allDevices == nil {
		t.allDevices = make(map[string]map[string]bool)
	}
	devices := make(map[string]bool)
	t.allDevices[userID] = devices

	for deviceID := range allDevices {
		devices[deviceID] = true
	}
}

func TestAddDeviceAfterRegister(t *testing.T) {
	trigger := newTestTrigger()

	var cfg service.Config
	err := envconfig.Process("", &cfg)
	require.NoError(t, err)
	cfg.Addr = "localhost:1234"

	shutdown := authService.NewAuthServer(t, cfg)
	defer shutdown()

	var acmeCfg certManager.Config
	err = envconfig.Process("LISTEN", &acmeCfg)
	require.NoError(t, err)
	certMgr, err := certManager.NewCertManager(acmeCfg)
	require.NoError(t, err)
	tlsConfig := certMgr.GetClientTLSConfig()

	conn, err := grpc.Dial(cfg.Addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	require.NoError(t, err)
	c := pb.NewAuthorizationServiceClient(conn)

	m := NewUserDevicesManager(trigger.Trigger, c, time.Millisecond*200, time.Millisecond*500, func(err error) { fmt.Println(err) })
	defer m.Close()
	err = m.Acquire(context.Background(), t.Name())
	require.NoError(t, err)

	_, err = c.AddDevice(context.Background(), &pb.AddDeviceRequest{
		UserId:   t.Name(),
		DeviceId: "deviceId_" + t.Name(),
	})

	time.Sleep(time.Second * 2)
	require.Equal(t, map[string]map[string]bool{
		t.Name(): {
			"deviceId_" + t.Name(): true,
		},
	}, trigger.Clone().allDevices)

	for i := 0; i < 5; i++ {
		devs, err := m.GetUserDevices(context.Background(), t.Name())
		require.NoError(t, err)
		require.NotEmpty(t, devs)
	}

	_, err = c.RemoveDevice(context.Background(), &pb.RemoveDeviceRequest{
		UserId:   t.Name(),
		DeviceId: "deviceId_" + t.Name(),
	})

	time.Sleep(time.Second * 2)
	require.Equal(t, map[string]map[string]bool(nil), trigger.Clone().allDevices)

	err = m.Release(t.Name())
	require.NoError(t, err)

	devs, err := m.GetUserDevices(context.Background(), t.Name())
	require.NoError(t, err)
	require.Empty(t, devs)

	_, err = c.AddDevice(context.Background(), &pb.AddDeviceRequest{
		UserId:   t.Name(),
		DeviceId: "deviceId_" + t.Name(),
	})
	time.Sleep(time.Second * 2)

	devs, err = m.GetUserDevices(context.Background(), t.Name())
	require.NoError(t, err)
	require.NotEmpty(t, devs)

	err = m.Release(t.Name())
	require.NoError(t, err)
}

func TestUserDevicesManager_Acquire(t *testing.T) {
	type fields struct {
		trigger *testTrigger
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *testTrigger
	}{
		{
			name: "empty - user not exist",
			fields: fields{
				trigger: newTestTrigger(),
			},
			args: args{
				userID: "notExist",
			},
			want: &testTrigger{},
		},
		{
			name: "valid",
			fields: fields{
				trigger: newTestTrigger(),
			},
			args: args{
				userID: t.Name(),
			},
			want: &testTrigger{
				addedDevices: map[string]map[string]bool{
					t.Name(): {
						"deviceId_" + t.Name(): true,
					},
				},
				allDevices: map[string]map[string]bool{
					t.Name(): {
						"deviceId_" + t.Name(): true,
					},
				},
			},
		},
	}

	var cfg service.Config
	err := envconfig.Process("", &cfg)
	require.NoError(t, err)
	cfg.Addr = "localhost:1234"

	shutdown := authService.NewAuthServer(t, cfg)
	defer shutdown()

	var acmeCfg certManager.Config
	err = envconfig.Process("LISTEN", &acmeCfg)
	require.NoError(t, err)
	certMgr, err := certManager.NewCertManager(acmeCfg)
	require.NoError(t, err)
	tlsConfig := certMgr.GetClientTLSConfig()

	conn, err := grpc.Dial(cfg.Addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	require.NoError(t, err)
	c := pb.NewAuthorizationServiceClient(conn)

	_, err = c.AddDevice(context.Background(), &pb.AddDeviceRequest{
		UserId:   t.Name(),
		DeviceId: "deviceId_" + t.Name(),
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewUserDevicesManager(tt.fields.trigger.Trigger, c, time.Millisecond*200, time.Second, func(err error) { fmt.Println(err) })
			defer m.Close()
			err := m.Acquire(context.Background(), tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				time.Sleep(time.Second)
				require.Equal(t, tt.want, tt.fields.trigger.Clone())
				err := m.Release(tt.args.userID)
				require.NoError(t, err)
			}
		})
	}
}

func TestUserDevicesManager_Release(t *testing.T) {
	type fields struct {
		trigger *testTrigger
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantErr      bool
		want         *testTrigger
		wantMgmtSize int
	}{
		{
			name: "empty - user not exist",
			fields: fields{
				trigger: newTestTrigger(),
			},
			args: args{
				userID: "notExist",
			},
			want: &testTrigger{},
		},
		{
			name: "valid",
			fields: fields{
				trigger: newTestTrigger(),
			},
			args: args{
				userID: t.Name(),
			},
			want: &testTrigger{
				addedDevices: map[string]map[string]bool{
					t.Name(): {
						"deviceId_" + t.Name(): true,
					},
				},
				removedDevices: map[string]map[string]bool{
					t.Name(): {
						"deviceId_" + t.Name(): true,
					},
				},
			},
			wantMgmtSize: 0,
		},
	}

	var cfg service.Config
	err := envconfig.Process("", &cfg)
	require.NoError(t, err)
	cfg.Addr = "localhost:1234"

	shutdown := authService.NewAuthServer(t, cfg)
	defer shutdown()

	var acmeCfg certManager.Config
	err = envconfig.Process("LISTEN", &acmeCfg)
	require.NoError(t, err)
	certMgr, err := certManager.NewCertManager(acmeCfg)
	require.NoError(t, err)
	tlsConfig := certMgr.GetClientTLSConfig()

	conn, err := grpc.Dial(cfg.Addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	require.NoError(t, err)
	c := pb.NewAuthorizationServiceClient(conn)

	_, err = c.AddDevice(context.Background(), &pb.AddDeviceRequest{
		UserId:   t.Name(),
		DeviceId: "deviceId_" + t.Name(),
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewUserDevicesManager(tt.fields.trigger.Trigger, c, time.Millisecond*200, time.Millisecond*500, func(err error) { fmt.Println(err) })
			defer m.Close()
			err := m.Acquire(context.Background(), tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				time.Sleep(time.Second)
				err := m.Release(tt.args.userID)
				require.NoError(t, err)
				require.Equal(t, tt.want, tt.fields.trigger.Clone())
				require.Equal(t, tt.wantMgmtSize, len(m.users))
			}
		})
	}
}
