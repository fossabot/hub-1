package service

import (
	"fmt"
	"time"

	"github.com/plgd-dev/hub/pkg/config"
	"github.com/plgd-dev/hub/pkg/log"
	"github.com/plgd-dev/hub/pkg/net/listener"
)

type AsymmetricKey struct {
	PrivateFile string
	PublicFile  string
}

type AccessTokenType string

const AccessTokenType_JWT AccessTokenType = "jwt"
const AccessTokenType_REFERENCE AccessTokenType = "reference"

type AllowedGrantType string

const AllowedGrantType_AUTHORIZATION_CODE AllowedGrantType = "authorization_code"
const AllowedGrantType_CLIENT_CREDENTIALS AllowedGrantType = "client_credentials"
const AllowedGrantType_PASSWORD AllowedGrantType = "password"
const AllowedGrantType_REFRESH_TOKEN AllowedGrantType = "refresh_token"

type AllowedGrantTypes []AllowedGrantType

func (gt AllowedGrantTypes) IsAllowed(v AllowedGrantType) bool {
	for _, t := range gt {
		if v == t {
			return true
		}
	}
	return false
}

type Client struct {
	ID                        string        `yaml:"id"`
	AuthorizationCodeLifetime time.Duration `yaml:"authorizationCodeLifetime"`
	AccessTokenLifetime       time.Duration `yaml:"accessTokenLifetime"`
	CodeRestrictionLifetime   time.Duration `yaml:"codeRestrictionLifetime"`
}

func (c *Client) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("id('%v')", c.ID)
	}
	return nil
}

type ClientsConfig []*Client

func (c ClientsConfig) Find(id string) *Client {
	for _, client := range c {
		if client.ID == id {
			return client
		}
	}
	return nil
}

// Config represents application configuration
type Config struct {
	Log         log.Config        `yaml:"log" json:"log"`
	APIs        APIsConfig        `yaml:"apis" json:"apis"`
	OAuthSigner OAuthSignerConfig `yaml:"oauthSigner" json:"oauthSigner"`
}

func (c *Config) Validate() error {
	if err := c.APIs.Validate(); err != nil {
		return fmt.Errorf("apis.%w", err)
	}
	if err := c.OAuthSigner.Validate(); err != nil {
		return fmt.Errorf("oauthSigner.%w", err)
	}
	return nil
}

// Config represent application configuration
type APIsConfig struct {
	HTTP listener.Config `yaml:"http" json:"http"`
}

func (c *APIsConfig) Validate() error {
	if err := c.HTTP.Validate(); err != nil {
		return fmt.Errorf("http.%w", err)
	}
	return nil
}

type OAuthSignerConfig struct {
	IDTokenKeyFile     string        `yaml:"idTokenKeyFile" json:"idTokenKeyFile"`
	AccessTokenKeyFile string        `yaml:"accessTokenKeyFile" json:"accessTokenKeyFile"`
	Domain             string        `yaml:"domain" json:"domain"`
	Clients            ClientsConfig `yaml:"clients" json:"clients"`
}

func (c *OAuthSignerConfig) Validate() error {
	if c.IDTokenKeyFile == "" {
		return fmt.Errorf("idTokenKeyFile('%v')", c.IDTokenKeyFile)
	}
	if c.AccessTokenKeyFile == "" {
		return fmt.Errorf("accessTokenKeyFile('%v')", c.AccessTokenKeyFile)
	}
	if c.Domain == "" {
		return fmt.Errorf("domain('%v')", c.Domain)
	}
	if len(c.Clients) == 0 {
		return fmt.Errorf("clients('%v')", c.Clients)
	}
	return nil
}

func (c Config) String() string {
	return config.ToString(c)
}
