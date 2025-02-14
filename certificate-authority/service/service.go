package service

import (
	"context"
	"fmt"

	"github.com/plgd-dev/hub/pkg/log"

	"github.com/plgd-dev/hub/pkg/net/grpc/server"
	"github.com/plgd-dev/hub/pkg/security/jwt/validator"
)

type Service struct {
	*server.Server
}

func New(ctx context.Context, config Config, logger log.Logger) (*Service, error) {
	validator, err := validator.New(ctx, config.APIs.GRPC.Authorization.Config, logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create validator: %w", err)
	}
	opts, err := server.MakeDefaultOptions(server.NewAuth(validator), logger)
	if err != nil {
		validator.Close()
		return nil, fmt.Errorf("cannot create grpc server options: %w", err)
	}
	server, err := server.New(config.APIs.GRPC, logger, opts...)

	if err != nil {
		validator.Close()
		return nil, err
	}
	server.AddCloseFunc(validator.Close)

	err = AddHandler(server, config.Signer)
	if err != nil {
		server.Close()
		return nil, err
	}

	return &Service{
		Server: server,
	}, nil
}
