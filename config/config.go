package config

import (
	"errors"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"log"
	"strings"
)

// Config is the central configuration type
type Config struct {
	Server   Server
	Webauthn WebauthnSettings
	Passcode Passcode
	Password Password
	Database Database
	Secrets  Secrets
	Cookies  Cookie
	Service  Service
}

func Load(cfgFile *string) (*Config, error) {
	k := koanf.New(".")
	var err error
	if cfgFile != nil && *cfgFile != "" {
		if err = k.Load(file.Provider(*cfgFile), yaml.Parser()); err == nil {
			log.Println("Using config file:", *cfgFile)
		} else {
			log.Println("Failed to load config from:", *cfgFile)
		}
	} else {
		if err = k.Load(file.Provider("./config/config.yaml"), yaml.Parser()); err == nil {
			log.Println("Using config file:", "./config/config.yaml")
		} else {
			log.Println("failed to load config from:", "./config/config.yaml")
		}
	}

	err = k.Load(env.Provider("", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "_", ".", -1)
	}), nil)
	if err != nil {
		log.Println("failed to load config from env vars")
	}

	c := defaultConfig()
	err = k.Unmarshal("", c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return c, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: Server{
			Public: ServerSettings{
				Address: ":8000",
			},
			Private: ServerSettings{
				Address: ":8001",
			},
		},
		Cookies: Cookie{
			HttpOnly: true,
			SameSite: "strict",
		},
		Webauthn: WebauthnSettings{
			RelyingParty: RelyingParty{
				Id:          "localhost",
				DisplayName: "Hanko Authentication Service",
				Origin:      "http://localhost",
			},
			Timeout: 60000,
		},
		Passcode: Passcode{
			Smtp: SMTP{
				Port: "465",
			},
			TTL: 300,
		},
		Database: Database{
			Database: "hanko",
		},
	}
}

func (c *Config) Validate() error {
	err := c.Server.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate server settings: %w", err)
	}
	err = c.Webauthn.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate webauthn settings: %w", err)
	}
	err = c.Passcode.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate passcode settings: %w", err)
	}
	err = c.Database.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate database settings: %w", err)
	}
	err = c.Secrets.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate secrets: %w", err)
	}
	err = c.Service.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate service settings: %w", err)
	}
	return nil
}

// Server contains the setting for the public and private server
type Server struct {
	Public  ServerSettings
	Private ServerSettings
}

func (s *Server) Validate() error {
	err := s.Public.Validate()
	if err != nil {
		return fmt.Errorf("error validating public server settings: %w", err)
	}
	err = s.Private.Validate()
	if err != nil {
		return fmt.Errorf("error validating private server settings: %w", err)
	}
	return nil
}

type Service struct {
	Name string
}

func (s *Service) Validate() error {
	if len(strings.TrimSpace(s.Name)) == 0 {
		return errors.New("field name must not be empty")
	}
	return nil
}

type Password struct {
	Enabled bool
}

type Cookie struct {
	Domain   string
	HttpOnly bool   `koanf:"http_only"`
	SameSite string `koanf:"same_site"`
}

type ServerSettings struct {
	// The Address to listen on in the form of host:port
	// See net.Dial for details of the address format.
	Address string
}

func (s *ServerSettings) Validate() error {
	if len(strings.TrimSpace(s.Address)) == 0 {
		return errors.New("field Address must not be empty")
	}
	return nil
}

// WebauthnSettings defines the settings for the webauthn authentication mechanism
type WebauthnSettings struct {
	RelyingParty RelyingParty `koanf:"relying_party"`
	Timeout      int
}

// Validate does not need to validate the config, because the library does this already
func (r *WebauthnSettings) Validate() error {
	return nil
}

// RelyingParty webauthn settings for your application using hanko.
type RelyingParty struct {
	Id          string
	DisplayName string `koanf:"display_name"`
	Icon        string
	Origin      string
}

// SMTP Server Settings for sending passcodes
type SMTP struct {
	Host     string
	Port     string
	User     string
	Password string
}

func (s *SMTP) Validate() error {
	if len(strings.TrimSpace(s.Host)) == 0 {
		return errors.New("smtp host must not be empty")
	}
	if len(strings.TrimSpace(s.Port)) == 0 {
		return errors.New("smtp port must not be empty")
	}
	if len(strings.TrimSpace(s.User)) == 0 {
		return errors.New("smtp user must not be empty")
	}
	if len(strings.TrimSpace(s.Password)) == 0 {
		return errors.New("smtp password must not be empty")
	}
	return nil
}

type Email struct {
	FromAddress string `koanf:"from_address"`
	FromName    string `koanf:"from_name"`
}

func (e *Email) Validate() error {
	if len(strings.TrimSpace(e.FromAddress)) == 0 {
		return errors.New("from_address must not be empty")
	}
	return nil
}

type Passcode struct {
	Email Email
	Smtp  SMTP
	TTL   int
}

func (p *Passcode) Validate() error {
	err := p.Email.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate email settings: %w", err)
	}
	err = p.Smtp.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate smtp settings: %w", err)
	}
	return nil
}

// Database connection settings
type Database struct {
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Dialect  string `json:"dialect"`
}

func (d *Database) Validate() error {
	if len(strings.TrimSpace(d.Database)) == 0 {
		return errors.New("database must not be empty")
	}
	if len(strings.TrimSpace(d.User)) == 0 {
		return errors.New("user must not be empty")
	}
	if len(strings.TrimSpace(d.Host)) == 0 {
		return errors.New("host must not be empty")
	}
	if len(strings.TrimSpace(d.Port)) == 0 {
		return errors.New("port must not be empty")
	}
	if len(strings.TrimSpace(d.Dialect)) == 0 {
		return errors.New("dialect must not be empty")
	}
	return nil
}

type Secrets struct {
	// Keys secret is used to en- and decrypt the JWKs which get used to sign the JWT tokens.
	// For every key a JWK is generated, encrypted with the key and persisted in the database.
	// The first key in the list is the one getting used for signing. If you want to use a new key, add it to the top of the list.
	// You can use this list for key rotation.
	// Each key must be at least 16 characters long.
	Keys []string `json:"keys"`
}

func (s *Secrets) Validate() error {
	if len(s.Keys) == 0 {
		return errors.New("at least one key must be defined")
	}
	return nil
}
