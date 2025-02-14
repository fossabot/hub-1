package test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/plgd-dev/hub/grpc-gateway/service"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/test/config"
	"github.com/stretchr/testify/require"
)

func MakeConfig(t *testing.T) service.Config {
	var cfg service.Config
	cfg.APIs.GRPC.Config = config.MakeGrpcServerConfig(config.GRPC_HOST)
	cfg.APIs.GRPC.OwnerCacheExpiration = time.Minute
	cfg.APIs.GRPC.SubscriptionBufferSize = 1000
	cfg.APIs.GRPC.TLS.ClientCertificateRequired = false

	cfg.Clients.IdentityStore.Connection = config.MakeGrpcClientConfig(config.IDENTITY_STORE_HOST)
	cfg.Clients.Eventbus.NATS = config.MakeSubscriberConfig()
	cfg.Clients.Eventbus.GoPoolSize = 16
	cfg.Clients.ResourceAggregate.Connection = config.MakeGrpcClientConfig(config.RESOURCE_AGGREGATE_HOST)
	cfg.Clients.ResourceDirectory.Connection = config.MakeGrpcClientConfig(config.RESOURCE_DIRECTORY_HOST)

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
		s.Close()
		wg.Wait()
	}
}
