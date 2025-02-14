package service

import (
	"fmt"
	"time"

	"github.com/plgd-dev/hub/grpc-gateway/pb"
	"github.com/plgd-dev/hub/pkg/config"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/pkg/net/grpc/client"
	"github.com/plgd-dev/hub/pkg/net/grpc/server"
	pkgTime "github.com/plgd-dev/hub/pkg/time"
	natsClient "github.com/plgd-dev/hub/resource-aggregate/cqrs/eventbus/nats/client"
	eventstoreConfig "github.com/plgd-dev/hub/resource-aggregate/cqrs/eventstore/config"
)

type Config struct {
	Log                     log.Config          `yaml:"log" json:"log"`
	APIs                    APIsConfig          `yaml:"apis" json:"apis"`
	Clients                 ClientsConfig       `yaml:"clients" json:"clients"`
	ExposedHubConfiguration PublicConfiguration `yaml:"publicConfiguration" json:"publicConfiguration"`
}

func (c *Config) Validate() error {
	if err := c.APIs.Validate(); err != nil {
		return fmt.Errorf("apis.%w", err)
	}
	if err := c.Clients.Validate(); err != nil {
		return fmt.Errorf("clients.%w", err)
	}
	if err := c.ExposedHubConfiguration.Validate(); err != nil {
		return fmt.Errorf("publicConfiguration.%w", err)
	}
	return nil
}

// Config represent application configuration
type APIsConfig struct {
	GRPC GRPCConfig `yaml:"grpc" json:"grpc"`
}

func (c *APIsConfig) Validate() error {
	if err := c.GRPC.Validate(); err != nil {
		return fmt.Errorf("grpc.%w", err)
	}
	return nil
}

type GRPCConfig struct {
	OwnerCacheExpiration time.Duration `yaml:"ownerCacheExpiration" json:"ownerCacheExpiration"`
	server.Config        `yaml:",inline" json:",inline"`
}

func (c *GRPCConfig) Validate() error {
	if c.OwnerCacheExpiration <= 0 {
		return fmt.Errorf("ownerCacheExpiration('%v')", c.OwnerCacheExpiration)
	}
	return c.Config.Validate()
}

type ClientsConfig struct {
	Eventbus      EventBusConfig      `yaml:"eventBus" json:"eventBus"`
	Eventstore    EventStoreConfig    `yaml:"eventStore" json:"eventStore"`
	IdentityStore IdentityStoreConfig `yaml:"identityStore" json:"identityStore"`
}

func (c *ClientsConfig) Validate() error {
	if err := c.IdentityStore.Validate(); err != nil {
		return fmt.Errorf("identityStore.%w", err)
	}
	if err := c.Eventbus.Validate(); err != nil {
		return fmt.Errorf("eventbus.%w", err)
	}
	if err := c.Eventstore.Validate(); err != nil {
		return fmt.Errorf("eventstore.%w", err)
	}
	return nil
}

type EventBusConfig struct {
	GoPoolSize int               `yaml:"goPoolSize" json:"goPoolSize"`
	NATS       natsClient.Config `yaml:"nats" json:"nats"`
}

func (c *EventBusConfig) Validate() error {
	if c.GoPoolSize <= 0 {
		return fmt.Errorf("goPoolSize('%v')", c.GoPoolSize)
	}
	if err := c.NATS.Validate(); err != nil {
		return fmt.Errorf("nats.%w", err)
	}
	return nil
}

type EventStoreConfig struct {
	ProjectionCacheExpiration time.Duration           `yaml:"cacheExpiration" json:"cacheExpiration"`
	Connection                eventstoreConfig.Config `yaml:",inline" json:",inline"`
}

func (c *EventStoreConfig) Validate() error {
	if c.ProjectionCacheExpiration <= 0 {
		return fmt.Errorf("cacheExpiration('%v')", c.ProjectionCacheExpiration)
	}
	return c.Connection.Validate()
}

type IdentityStoreConfig struct {
	Connection client.Config `yaml:"grpc" json:"grpc"`
}

func (c *IdentityStoreConfig) Validate() error {
	if err := c.Connection.Validate(); err != nil {
		return fmt.Errorf("grpc.%w", err)
	}
	return nil
}

type PublicConfiguration struct {
	CAPool                   string        `yaml:"caPool" json:"caPool" description:"file path to the root certificate in PEM format"`
	OwnerClaim               string        `yaml:"ownerClaim" json:"ownerClaim"`
	DeviceIDClaim            string        `yaml:"deviceIdClaim" json:"deviceIdClaim"`
	HubID                    string        `yaml:"hubId" json:"hubId"`
	CoapGateway              string        `yaml:"coapGateway" json:"coapGateway"`
	DefaultCommandTimeToLive time.Duration `yaml:"defaultCommandTimeToLive" json:"defaultCommandTimeToLive"`
	AuthorizationServer      string        `yaml:"authorizationServer" json:"authorizationServer"`

	cloudCertificateAuthorities string `yaml:"-"`
}

func (c *PublicConfiguration) Validate() error {
	if c.CAPool == "" {
		return fmt.Errorf("caPool('%v')", c.CAPool)
	}
	if c.OwnerClaim == "" {
		return fmt.Errorf("ownerClaim('%v')", c.OwnerClaim)
	}
	if c.HubID == "" {
		return fmt.Errorf("hubId('%v')", c.HubID)
	}
	if c.CoapGateway == "" {
		return fmt.Errorf("coapGateway('%v')", c.CoapGateway)
	}
	if c.CAPool == "" {
		return fmt.Errorf("caPool('%v')", c.CAPool)
	}
	if c.AuthorizationServer == "" {
		return fmt.Errorf("authorizationServer('%v')", c.AuthorizationServer)
	}
	return nil
}

func (c PublicConfiguration) ToProto() *pb.HubConfigurationResponse {
	return &pb.HubConfigurationResponse{
		JwtOwnerClaim:            c.OwnerClaim,
		JwtDeviceIdClaim:         c.DeviceIDClaim,
		Id:                       c.HubID,
		CoapGateway:              c.CoapGateway,
		CertificateAuthorities:   c.cloudCertificateAuthorities,
		DefaultCommandTimeToLive: int64(c.DefaultCommandTimeToLive),
		CurrentTime:              pkgTime.UnixNano(time.Now()),
		AuthorizationServer:      c.AuthorizationServer,
	}
}

//String return string representation of Config
func (c Config) String() string {
	return config.ToString(c)
}
