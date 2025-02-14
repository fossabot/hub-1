package service

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/plgd-dev/hub/identity-store/persistence"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus/nats/test"
	"github.com/plgd-dev/hub/test/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID   = "testUserID"
	testDeviceID = "testDeviceID"
	testUser2    = "testUser2"
)

func makeConfig(t *testing.T) Config {
	var cfg Config

	cfg.APIs.GRPC.Addr = config.IDENTITY_STORE_HOST
	cfg.APIs.GRPC.TLS.CAPool = config.CA_POOL
	cfg.APIs.GRPC.TLS.CertFile = config.CERT_FILE
	cfg.APIs.GRPC.TLS.KeyFile = config.KEY_FILE
	cfg.APIs.GRPC.Authorization.OwnerClaim = config.OWNER_CLAIM
	cfg.APIs.GRPC.Authorization.Config = config.MakeAuthorizationConfig()

	cfg.Clients.Storage.MongoDB.URI = config.MONGODB_URI
	cfg.Clients.Storage.MongoDB.Database = config.IDENTITY_STORE_DB
	cfg.Clients.Storage.MongoDB.TLS.CAPool = config.CA_POOL
	cfg.Clients.Storage.MongoDB.TLS.CertFile = config.CERT_FILE
	cfg.Clients.Storage.MongoDB.TLS.KeyFile = config.KEY_FILE

	cfg.Clients.Eventbus.NATS = config.MakePublisherConfig()

	err := cfg.Validate()
	require.NoError(t, err)

	return cfg
}

func newTestService(t *testing.T) (*Server, func()) {
	cfg := makeConfig(t)

	logger, err := log.NewLogger(cfg.Log)
	require.NoError(t, err)

	naClient, publisher, err := test.NewClientAndPublisher(cfg.Clients.Eventbus.NATS, logger)
	require.NoError(t, err)

	s, err := NewServer(context.Background(), cfg, logger, publisher)
	require.NoError(t, err)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_ = s.Serve()
		defer wg.Done()
	}()
	return s, func() {
		s.Shutdown()
		publisher.Close()
		naClient.Close()
		wg.Wait()
	}
}

func (s *Server) cleanUp() error {
	p := s.service.persistence
	var errors []error
	if err := p.Clear(context.Background()); err != nil {
		errors = append(errors, err)
	}
	if err := p.Close(context.Background()); err != nil {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return fmt.Errorf("%v", errors)
	}
	return nil
}

func newTestDevice() *persistence.AuthorizedDevice {
	return newTestDeviceWithIDAndOwner(testDeviceID, testUserID)
}

func newTestDeviceWithIDAndOwner(deviceID, owner string) *persistence.AuthorizedDevice {
	return &persistence.AuthorizedDevice{
		DeviceID: deviceID,
		Owner:    owner,
	}
}

func persistDevice(t *testing.T, p Persistence, d *persistence.AuthorizedDevice) {
	tx := p.NewTransaction(context.Background())
	defer tx.Close()
	err := tx.Persist(d)
	assert.Nil(t, err)
}
