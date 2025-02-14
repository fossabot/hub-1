package client

import (
	"fmt"

	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/pkg/security/certManager/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// GRPC Client.
type Client struct {
	client    *grpc.ClientConn
	closeFunc []func()
}

func (c *Client) GRPC() *grpc.ClientConn {
	return c.client
}

// AddCloseFunc adds a function to be called by the Close method.
// This eliminates the need for wrapping the Client.
func (s *Client) AddCloseFunc(f func()) {
	s.closeFunc = append(s.closeFunc, f)
}

func (s *Client) Close() error {
	err := s.client.Close()
	for _, f := range s.closeFunc {
		f()
	}
	return err
}

func New(config Config, logger log.Logger, opts ...grpc.DialOption) (*Client, error) {
	certManager, err := client.New(config.TLS, logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create cert manager: %w", err)
	}
	v := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(certManager.GetTLSConfig())),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.KeepAlive.Time,
			Timeout:             config.KeepAlive.Timeout,
			PermitWithoutStream: config.KeepAlive.PermitWithoutStream,
		}),
	}
	if len(opts) > 0 {
		v = append(v, opts...)
	}

	conn, err := grpc.Dial(config.Addr, v...)
	if err != nil {
		return nil, fmt.Errorf("cannot dial: %w", err)
	}

	return &Client{
		client: conn, closeFunc: []func(){certManager.Close},
	}, nil
}
