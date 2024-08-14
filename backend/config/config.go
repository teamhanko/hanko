package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/teamhanko/hanko/backend/ee/saml/config"
	"log"
)

// Config is the central configuration type
type Config struct {
	// `account` configures settings related to user accounts.
	Account Account `yaml:"account" json:"account,omitempty" koanf:"account" jsonschema:"title=account"`
	// `audit_log` configures output and storage modalities of audit logs.
	AuditLog AuditLog `yaml:"audit_log" json:"audit_log,omitempty" koanf:"audit_log" split_words:"true" jsonschema:"title=audit_log"`
	// `convert_legacy_config`, if set to `true`, automatically copies the set values of deprecated configuration
	// options, to new ones. If set to `false`, these values have to be set manually if non-default values should be
	// used.
	ConvertLegacyConfig bool `yaml:"convert_legacy_config" json:"convert_legacy_config,omitempty" koanf:"convert_legacy_config" split_words:"true" jsonschema:"default=false"`
	// `database` configures database connection settings.
	Database Database `yaml:"database" json:"database,omitempty" koanf:"database" jsonschema:"title=database"`
	// `debug`, if set to `true`, adds additional debugging information to flow API responses.
	Debug bool `yaml:"debug" json:"debug,omitempty" koanf:"debug" jsonschema:"default=false"`
	// `email` configures how email addresses of user accounts are acquired and used.
	Email Email `yaml:"email" json:"email,omitempty" koanf:"email" jsonschema:"title=email"`
	// `email_delivery` configures how outgoing mails are delivered.
	EmailDelivery EmailDelivery `yaml:"email_delivery" json:"email_delivery,omitempty" koanf:"email_delivery" split_words:"true" jsonschema:"title=email_delivery"`
	// Deprecated. See child properties for suggested replacements.
	Emails Emails `yaml:"emails" json:"emails,omitempty" koanf:"emails" jsonschema:"title=emails"`
	// `log` configures application logging.
	Log LoggerConfig `yaml:"log" json:"log,omitempty" koanf:"log" jsonschema:"title=log"`
	// Deprecated. See child properties for suggested replacements.
	Passcode Passcode `yaml:"passcode" json:"passcode,omitempty" koanf:"passcode" jsonschema:"title=passcode"`
	// `passkey` configures how passkeys  are acquired and used.
	Passkey Passkey `yaml:"passkey" json:"passkey,omitempty" koanf:"passkey" jsonschema:"title=passkey"`
	// `password` configures how passwords are acquired and used.
	Password Password `yaml:"password" json:"password,omitempty" koanf:"password" jsonschema:"title=password"`
	// `rate_limiter` configures rate limits for rate limited API operations and storage modalities for rate limit data.
	RateLimiter RateLimiter `yaml:"rate_limiter" json:"rate_limiter,omitempty" koanf:"rate_limiter" split_words:"true" jsonschema:"title=rate_limiter"`
	// `saml` configures modalities of SAML (Security Assertion Markup Language) SSO authentication and SAML identity
	// providers.
	Saml config.Saml `yaml:"saml" json:"saml,omitempty" koanf:"saml" jsonschema:"title=saml"`
	// `secrets` configures the keys used for cryptographically signing tokens issued by the API.
	Secrets Secrets `yaml:"secrets" json:"secrets,omitempty" koanf:"secrets" jsonschema:"title=secrets"`
	// `server` configures address and CORS settings of the public and admin API.
	Server Server `yaml:"server" json:"server,omitempty" koanf:"server" jsonschema:"title=server"`
	// `service` configures general service information.
	Service Service `yaml:"service" json:"service,omitempty" koanf:"service" jsonschema:"title=service"`
	// `session` configures settings for session JWTs and Cookies issued by the API.
	Session Session `yaml:"session" json:"session,omitempty" koanf:"session" jsonschema:"title=session"`
	// Deprecated. Use `email_delivery.smtp` instead.
	Smtp SMTP `yaml:"smtp" json:"smtp,omitempty" koanf:"smtp" jsonschema:"title=smtp"`
	// `third_party` configures the modalities of third party OAuth/OIDC based authentication and available identity
	// providers.
	ThirdParty ThirdParty `yaml:"third_party" json:"third_party,omitempty" koanf:"third_party" split_words:"true" jsonschema:"title=third_party"`
	// `username` configures how usernames of user accounts are acquired and used.
	Username Username `yaml:"username" json:"username,omitempty" koanf:"username" jsonschema:"title=username"`
	// `webauthn` configures general settings for communication with the WebAuthentication API.
	Webauthn WebauthnSettings `yaml:"webauthn" json:"webauthn,omitempty" koanf:"webauthn" jsonschema:"title=webauthn"`
	// `webhooks` configures HTTP-based callbacks for specific events occurring in the system.
	Webhooks WebhookSettings `yaml:"webhooks" json:"webhooks,omitempty" koanf:"webhooks" jsonschema:"title=webhooks"`
}

var (
	DefaultConfigFilePath = "./config/config.yaml"
)

func LoadFile(filePath *string, pa koanf.Parser) (*koanf.Koanf, error) {
	k := koanf.New(".")

	if filePath == nil || *filePath == "" {
		return nil, nil
	}

	if err := k.Load(file.Provider(*filePath), pa); err != nil {
		return nil, fmt.Errorf("failed to load file from '%s': %w", *filePath, err)
	}

	return k, nil
}

func Load(cfgFile *string) (*Config, error) {
	if cfgFile == nil || *cfgFile == "" {
		*cfgFile = DefaultConfigFilePath
	}

	k, err := LoadFile(cfgFile, yaml.Parser())
	if err != nil {
		if *cfgFile != DefaultConfigFilePath {
			return nil, fmt.Errorf("failed to load config from: %s: %w", *cfgFile, err)
		}
		log.Println("failed to load config, skipping...")
	} else {
		log.Println("Using config file:", *cfgFile)
	}

	c := DefaultConfig()
	err = k.Unmarshal("", c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	err = envconfig.Process("", c)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from env vars: %w", err)
	}

	err = c.PostProcess()
	if err != nil {
		return nil, fmt.Errorf("failed to post process config: %w", err)
	}

	if err = c.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %s", err)
	}

	return c, nil
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
	if c.EmailDelivery.Enabled {
		err = c.Smtp.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate smtp settings: %w", err)
		}
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
	err = c.Session.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate session settings: %w", err)
	}
	err = c.RateLimiter.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate rate-limiter settings: %w", err)
	}
	err = c.ThirdParty.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate third_party settings: %w", err)
	}
	err = c.Saml.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate saml settings: %w", err)
	}
	err = c.Webhooks.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate webhook settings: %w", err)
	}
	return nil
}

func (c *Config) convertLegacyConfig() {
	c.Email.Limit = c.Emails.MaxNumOfAddresses
	c.Email.RequireVerification = c.Emails.RequireVerification
	c.Email.PasscodeTtl = c.Passcode.TTL

	c.EmailDelivery.SMTP = c.Smtp

	c.Password.MinLength = c.Password.MinPasswordLength

	c.Passkey.UserVerification = c.Webauthn.UserVerification

	c.Webauthn.Timeouts.Login = c.Webauthn.Timeout
	c.Webauthn.Timeouts.Registration = c.Webauthn.Timeout
}

func (c *Config) PostProcess() error {
	if c.ConvertLegacyConfig {
		c.convertLegacyConfig()
	}

	err := c.ThirdParty.PostProcess()
	if err != nil {
		return fmt.Errorf("failed to post process third party settings: %w", err)
	}

	err = c.Webauthn.PostProcess()
	if err != nil {
		return fmt.Errorf("failed to post process webauthn settings: %w", err)
	}

	err = c.Saml.PostProcess()
	if err != nil {
		return fmt.Errorf("failed to post process saml settings: %w", err)
	}

	return nil
}
