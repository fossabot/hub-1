package test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/resource-aggregate/service"
	"github.com/plgd-dev/hub/test/config"
	"github.com/stretchr/testify/require"
)

func MakeConfig(t *testing.T) service.Config {
	var cfg service.Config

	cfg.APIs.GRPC.OwnerCacheExpiration = time.Minute
	cfg.APIs.GRPC.Config = config.MakeGrpcServerConfig(config.RESOURCE_AGGREGATE_HOST)

	cfg.Clients.IdentityStore.Connection = config.MakeGrpcClientConfig(config.IDENTITY_STORE_HOST)

	cfg.Clients.Eventbus.NATS = config.MakePublisherConfig()

	cfg.Clients.Eventstore.Connection.MongoDB = config.MakeEventsStoreMongoDBConfig()
	cfg.Clients.Eventstore.ConcurrencyExceptionMaxRetry = 8
	cfg.Clients.Eventstore.SnapshotThreshold = 16

	err := cfg.Validate()
	require.NoError(t, err)

	return cfg
}

func SetUp(t *testing.T) (TearDown func()) {
	return New(t, MakeConfig(t))
}

func New(t *testing.T, cfg service.Config) func() {
	ctx := context.Background()
	logger, err := log.NewLogger(cfg.Log)
	require.NoError(t, err)

	s, err := service.New(ctx, cfg, logger)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = s.Serve()
	}()

	return func() {
		s.Shutdown()
		wg.Wait()
	}
}
