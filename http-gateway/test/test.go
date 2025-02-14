package test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/plgd-dev/hub/grpc-gateway/pb"
	"github.com/plgd-dev/hub/http-gateway/service"
	"github.com/plgd-dev/hub/http-gateway/uri"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/test/config"
	"github.com/plgd-dev/kit/v2/codec/cbor"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func MakeConfig(t *testing.T) service.Config {
	var cfg service.Config
	cfg.APIs.HTTP.Authorization = config.MakeAuthorizationConfig()
	cfg.APIs.HTTP.Connection = config.MakeListenerConfig(config.HTTP_GW_HOST)
	cfg.APIs.HTTP.Connection.TLS.ClientCertificateRequired = false
	cfg.APIs.HTTP.WebSocket.StreamBodyLimit = 256 * 1024
	cfg.APIs.HTTP.WebSocket.PingFrequency = 10 * time.Second

	cfg.Clients.GrpcGateway.Connection = config.MakeGrpcClientConfig(config.GRPC_HOST)

	err := cfg.Validate()
	require.NoError(t, err)

	fmt.Printf("cfg\n%v\n", cfg.String())

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
		_ = s.Shutdown()
		wg.Wait()
	}
}

func GetContentData(content *pb.Content, desiredContentType string) ([]byte, error) {
	if desiredContentType == uri.ApplicationProtoJsonContentType {
		data, err := protojson.Marshal(content)
		if err != nil {
			return nil, err
		}
		return data, err
	}
	v, err := cbor.ToJSON(content.GetData())
	if err != nil {
		return nil, err
	}
	return []byte(v), err
}
