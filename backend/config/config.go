package config

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobwas/glob"
	"github.com/kelseyhightower/envconfig"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/ee/saml/config"
	"golang.org/x/exp/slices"
)

// Config is the central configuration type
type Config struct {
	ConvertLegacyConfig bool             `yaml:"convert_legacy_config" json:"convert_legacy_config,omitempty" koanf:"convert_legacy_config" split_words:"true"`
	Server              Server           `yaml:"server" json:"server,omitempty" koanf:"server"`
	Webauthn            WebauthnSettings `yaml:"webauthn" json:"webauthn,omitempty" koanf:"webauthn"`
	Smtp                SMTP             `yaml:"smtp" json:"smtp,omitempty" koanf:"smtp"` // Deprecated, use EmailDelivery.SMTP instead
	EmailDelivery       EmailDelivery    `yaml:"email_delivery" json:"email_delivery,omitempty" koanf:"email_delivery" split_words:"true"`
	Passcode            Passcode         `yaml:"passcode" json:"passcode,omitempty" koanf:"passcode"` // Deprecated
	Password            Password         `yaml:"password" json:"password,omitempty" koanf:"password"`
	Database            Database         `yaml:"database" json:"database,omitempty" koanf:"database"`
	Secrets             Secrets          `yaml:"secrets" json:"secrets,omitempty" koanf:"secrets"`
	Service             Service          `yaml:"service" json:"service,omitempty" koanf:"service"`
	Session             Session          `yaml:"session" json:"session,omitempty" koanf:"session"`
	AuditLog            AuditLog         `yaml:"audit_log" json:"audit_log,omitempty" koanf:"audit_log" split_words:"true"`
	Emails              Emails           `yaml:"emails" json:"emails,omitempty" koanf:"emails"`
	RateLimiter         RateLimiter      `yaml:"rate_limiter" json:"rate_limiter,omitempty" koanf:"rate_limiter" split_words:"true"`
	ThirdParty          ThirdParty       `yaml:"third_party" json:"third_party,omitempty" koanf:"third_party" split_words:"true"`
	Log                 LoggerConfig     `yaml:"log" json:"log,omitempty" koanf:"log"`
	Account             Account          `yaml:"account" json:"account,omitempty" koanf:"account"`
	SecondFactor        SecondFactor     `yaml:"second_factor" json:"second_factor,omitempty" koanf:"second_factor" split_word:"true"`
	Passkey             Passkey          `yaml:"passkey" json:"passkey,omitempty" koanf:"passkey"`
	Saml                config.Saml      `yaml:"saml" json:"saml,omitempty" koanf:"saml"`
	Webhooks            WebhookSettings  `yaml:"webhooks" json:"webhooks,omitempty" koanf:"webhooks"`
	Email               Email            `yaml:"email" json:"email,omitempty" koanf:"email"`
	Username            Username         `yaml:"username" json:"username,omitempty" koanf:"username"`
	Debug               bool             `yaml:"debug" json:"debug,omitempty" koanf:"debug"`
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

	if err := envconfig.Process("", c); err != nil {
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
	if c.EmailDelivery.Enabled {
		err = c.Smtp.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate smtp settings: %w", err)
		}
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

// Server contains the setting for the public and admin server
type Server struct {
	Public ServerSettings `yaml:"public" json:"public,omitempty" koanf:"public"`
	Admin  ServerSettings `yaml:"admin" json:"admin,omitempty" koanf:"admin"`
}

func (s *Server) Validate() error {
	err := s.Public.Validate()
	if err != nil {
		return fmt.Errorf("error validating public server settings: %w", err)
	}
	err = s.Admin.Validate()
	if err != nil {
		return fmt.Errorf("error validating admin server settings: %w", err)
	}
	return nil
}

type Service struct {
	Name string `yaml:"name" json:"name,omitempty" koanf:"name"`
}

func (s *Service) Validate() error {
	if len(strings.TrimSpace(s.Name)) == 0 {
		return errors.New("field name must not be empty")
	}
	return nil
}

type Password struct {
	Enabled               bool   `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	Optional              bool   `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=false"`
	AcquireOnRegistration string `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=never,enum=always,enum=conditional,enum=never"`
	AcquireOnLogin        string `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=always,enum=always,enum=conditional,enum=never"`
	Recovery              bool   `yaml:"recovery" json:"recovery,omitempty" koanf:"recovery" jsonschema:"default=true"`
	MinLength             int    `yaml:"min_length" json:"min_length,omitempty" koanf:"min_length" split_words:"true" jsonschema:"default=8"`
}

type Cookie struct {
	Name     string `yaml:"name" json:"name,omitempty" koanf:"name" jsonschema:"default=hanko"`
	Domain   string `yaml:"domain" json:"domain,omitempty" koanf:"domain"`
	HttpOnly bool   `yaml:"http_only" json:"http_only,omitempty" koanf:"http_only" split_words:"true"`
	SameSite string `yaml:"same_site" json:"same_site,omitempty" koanf:"same_site" split_words:"true"`
	Secure   bool   `yaml:"secure" json:"secure,omitempty" koanf:"secure"`
}

func (c *Cookie) GetName() string {
	if c.Name != "" {
		return c.Name
	}

	return "hanko"
}

type ServerSettings struct {
	// The Address to listen on in the form of host:port
	// See net.Dial for details of the address format.
	Address string `yaml:"address" json:"address,omitempty" koanf:"address"`
	Cors    Cors   `yaml:"cors" json:"cors,omitempty" koanf:"cors"`
}

type Cors struct {
	// AllowOrigins determines the value of the Access-Control-Allow-Origin
	// response header. This header defines a list of origins that may access the
	// resource.  The wildcard characters '*' and '?' are supported and are
	// converted to regex fragments '.*' and '.' accordingly.
	AllowOrigins []string `yaml:"allow_origins" json:"allow_origins,omitempty" koanf:"allow_origins" split_words:"true"`

	// UnsafeWildcardOriginWithAllowCredentials UNSAFE/INSECURE: allows wildcard '*' origin to be used with AllowCredentials
	// flag. In that case we consider any origin allowed and send it back to the client with `Access-Control-Allow-Origin` header.
	//
	// This is INSECURE and potentially leads to [cross-origin](https://portswigger.net/research/exploiting-cors-misconfigurations-for-bitcoins-and-bounties)
	// attacks. See: https://github.com/labstack/echo/issues/2400 for discussion on the subject.
	//
	// Optional. Default value is false.
	UnsafeWildcardOriginAllowed bool `yaml:"unsafe_wildcard_origin_allowed" json:"unsafe_wildcard_origin_allowed,omitempty" koanf:"unsafe_wildcard_origin_allowed" split_words:"true" jsonschema:"default=false"`
}

func (cors *Cors) Validate() error {
	for _, origin := range cors.AllowOrigins {
		if origin == "*" && !cors.UnsafeWildcardOriginAllowed {
			return fmt.Errorf("found wildcard '*' origin in server.public.cors.allow_origins, if this is intentional set server.public.cors.unsafe_wildcard_origin_allowed to true")
		}
	}

	return nil
}

func (s *ServerSettings) Validate() error {
	if len(strings.TrimSpace(s.Address)) == 0 {
		return errors.New("field Address must not be empty")
	}
	if err := s.Cors.Validate(); err != nil {
		return err
	}
	return nil
}

type WebauthnTimeouts struct {
	Registration int `yaml:"registration" json:"registration,omitempty" koanf:"registration"`
	Login        int `yaml:"login" json:"login,omitempty" koanf:"login"`
}

// WebauthnSettings defines the settings for the webauthn authentication mechanism
type WebauthnSettings struct {
	RelyingParty RelyingParty `yaml:"relying_party" json:"relying_party,omitempty" koanf:"relying_party" split_words:"true"`
	// Deprecated, use Timeouts instead
	Timeout  int              `yaml:"timeout" json:"timeout,omitempty" koanf:"timeout" jsonschema:"default=60000"`
	Timeouts WebauthnTimeouts `yaml:"timeouts" json:"timeouts,omitempty" koanf:"timeouts" split_words:"true"`
	// Deprecated, use Passkey.UserVerification instead
	UserVerification string                `yaml:"user_verification" json:"user_verification,omitempty" koanf:"user_verification" split_words:"true" jsonschema:"default=preferred,enum=required,enum=preferred,enum=discouraged"`
	Handler          *webauthnLib.WebAuthn `jsonschema:"-"`
}

func (r *WebauthnSettings) PostProcess() error {
	requireResidentKey := false

	config := &webauthnLib.Config{
		RPID:                  r.RelyingParty.Id,
		RPDisplayName:         r.RelyingParty.DisplayName,
		RPOrigins:             r.RelyingParty.Origins,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &requireResidentKey,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Debug: false,
		Timeouts: webauthnLib.TimeoutsConfig{
			Login: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(r.Timeouts.Login) * time.Millisecond,
			},
			Registration: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(r.Timeouts.Registration) * time.Millisecond,
			},
		},
	}

	handler, err := webauthnLib.New(config)
	if err != nil {
		return err
	}

	r.Handler = handler

	return nil
}

// Validate does not need to validate the config, because the library does this already
func (r *WebauthnSettings) Validate() error {
	validUv := []string{"required", "preferred", "discouraged"}
	if !slices.Contains(validUv, r.UserVerification) {
		return fmt.Errorf("expected user_verification to be one of [%s], got: '%s'", strings.Join(validUv, ", "), r.UserVerification)
	}
	return nil
}

// RelyingParty webauthn settings for your application using hanko.
type RelyingParty struct {
	Id          string   `yaml:"id" json:"id,omitempty" koanf:"id" jsonschema:"default=localhost"`
	DisplayName string   `yaml:"display_name" json:"display_name,omitempty" koanf:"display_name" split_words:"true" jsonschema:"default=Hanko Authentication Service"`
	Icon        string   `yaml:"icon" json:"icon,omitempty" koanf:"icon"`
	Origins     []string `yaml:"origins" json:"origins,omitempty" koanf:"origins" jsonschema:"minItems=1,default=http://localhost:8888"`
}

// SMTP Server Settings for sending passcodes
type SMTP struct {
	Host     string `yaml:"host" json:"host,omitempty" koanf:"host"`
	Port     string `yaml:"port" json:"port,omitempty" koanf:"port" jsonschema:"default=465,oneof_type=string;integer"`
	User     string `yaml:"user" json:"user,omitempty" koanf:"user"`
	Password string `yaml:"password" json:"password,omitempty" koanf:"password"`
}

func (s *SMTP) Validate() error {
	if len(strings.TrimSpace(s.Host)) == 0 {
		return errors.New("smtp host must not be empty")
	}
	if len(strings.TrimSpace(s.Port)) == 0 {
		return errors.New("smtp port must not be empty")
	}
	return nil
}

type PasscodeEmail struct {
	FromAddress string `yaml:"from_address" json:"from_address,omitempty" koanf:"from_address" split_words:"true" jsonschema:"default=passcode@hanko.io"`
	FromName    string `yaml:"from_name" json:"from_name,omitempty" koanf:"from_name" split_words:"true" jsonschema:"default=Hanko"`
}

func (e *PasscodeEmail) Validate() error {
	if len(strings.TrimSpace(e.FromAddress)) == 0 {
		return errors.New("from_address must not be empty")
	}
	return nil
}

type Passcode struct {
	Email PasscodeEmail `yaml:"email" json:"email,omitempty" koanf:"email"`
	TTL   int           `yaml:"ttl" json:"ttl,omitempty" koanf:"ttl" jsonschema:"default=300"`
	//Deprecated: Use root level Smtp instead
	Smtp SMTP `yaml:"smtp" json:"smtp,omitempty" koanf:"smtp,omitempty" required:"false" envconfig:"smtp,omitempty"`
}

func (p *Passcode) Validate() error {
	err := p.Email.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate email settings: %w", err)
	}
	return nil
}

// Database connection settings
type Database struct {
	Database string `yaml:"database" json:"database,omitempty" koanf:"database" jsonschema:"default=hanko"`
	User     string `yaml:"user" json:"user,omitempty" koanf:"user" jsonschema:"default=hanko"`
	Password string `yaml:"password" json:"password,omitempty" koanf:"password" jsonschema:"default=hanko"`
	Host     string `yaml:"host" json:"host,omitempty" koanf:"host" jsonschema:"default=localhost"`
	Port     string `yaml:"port" json:"port,omitempty" koanf:"port" jsonschema:"default=\"5432\""`
	Dialect  string `yaml:"dialect" json:"dialect,omitempty" koanf:"dialect" jsonschema:"default=postgres,enum=postgres,enum=mysql,enum=cockroach"`
	Url      string `yaml:"url" json:"url,omitempty" koanf:"url"`
}

func (d *Database) Validate() error {
	if len(strings.TrimSpace(d.Url)) > 0 {
		return nil
	}
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
	// Keys secrets are used to en- and decrypt the JWKs which get used to sign the JWTs.
	// For every key a JWK is generated, encrypted with the key and persisted in the database.
	//
	// You can use this list for key rotation: add a new key to the beginning of the list and the corresponding
	// JWK will then be used for signing JWTs. All tokens signed with the previous JWK(s) will still
	// be valid until they expire. Removing a key from the list does not remove the corresponding
	// database record. If you remove a key, you also have to remove the database record, otherwise
	// application startup will fail.
	//
	// Each key must be at least 16 characters long.
	Keys []string `yaml:"keys" json:"keys,omitempty" koanf:"keys" jsonschema:"minItems=1"`
}

func (s *Secrets) Validate() error {
	if len(s.Keys) == 0 {
		return errors.New("at least one key must be defined")
	}
	return nil
}

type Session struct {
	EnableAuthTokenHeader bool `yaml:"enable_auth_token_header" json:"enable_auth_token_header,omitempty" koanf:"enable_auth_token_header" split_words:"true" jsonschema:"default=false"`
	// Lifespan, possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Lifespan string `yaml:"lifespan" json:"lifespan,omitempty" koanf:"lifespan" jsonschema:"default=1h"`
	Cookie   Cookie `yaml:"cookie" json:"cookie,omitempty" koanf:"cookie"`

	// Issuer optional string to be used in the jwt iss claim.
	Issuer string `yaml:"issuer" json:"issuer,omitempty" koanf:"issuer"`

	// Audience optional []string containing strings which get put into the aud claim. If not set default to Webauthn.RelyingParty.Id config parameter.
	Audience []string `yaml:"audience" json:"audience,omitempty" koanf:"audience"`
}

func (s *Session) Validate() error {
	_, err := time.ParseDuration(s.Lifespan)
	if err != nil {
		return errors.New("failed to parse lifespan")
	}
	return nil
}

type AuditLog struct {
	ConsoleOutput AuditLogConsole `yaml:"console_output" json:"console_output,omitempty" koanf:"console_output" split_words:"true"`
	Storage       AuditLogStorage `yaml:"storage" json:"storage,omitempty" koanf:"storage"`
	Mask          bool            `yaml:"mask" json:"mask,omitempty" koanf:"mask" jsonschema:"default=true"`
}

type AuditLogStorage struct {
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
}

type AuditLogConsole struct {
	Enabled      bool         `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	OutputStream OutputStream `yaml:"output" json:"output,omitempty" koanf:"output" split_words:"true" jsonschema:"default=stdout,enum=stdout,enum=stderr"`
}

type Emails struct {
	RequireVerification bool `yaml:"require_verification" json:"require_verification,omitempty" koanf:"require_verification" split_words:"true" jsonschema:"default=true"`
	MaxNumOfAddresses   int  `yaml:"max_num_of_addresses" json:"max_num_of_addresses,omitempty" koanf:"max_num_of_addresses" split_words:"true" jsonschema:"default=5"`
}

type OutputStream string

var (
	OutputStreamStdOut OutputStream = "stdout"
	OutputStreamStdErr OutputStream = "stderr"
)

type RateLimiter struct {
	Enabled        bool                 `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	Store          RateLimiterStoreType `yaml:"store" json:"store,omitempty" koanf:"store" jsonschema:"default=in_memory,enum=in_memory,enum=redis"`
	Redis          *RedisConfig         `yaml:"redis_config" json:"redis_config,omitempty" koanf:"redis_config"`
	PasscodeLimits RateLimits           `yaml:"passcode_limits" json:"passcode_limits,omitempty" koanf:"passcode_limits" split_words:"true"`
	PasswordLimits RateLimits           `yaml:"password_limits" json:"password_limits,omitempty" koanf:"password_limits" split_words:"true"`
	TokenLimits    RateLimits           `yaml:"token_limits" json:"token_limits,omitempty" koanf:"token_limits" split_words:"true"`
}

type RateLimits struct {
	Tokens   uint64        `yaml:"tokens" json:"tokens" koanf:"tokens"`
	Interval time.Duration `yaml:"interval" json:"interval" koanf:"interval"`
}

type RateLimiterStoreType string

const (
	RATE_LIMITER_STORE_IN_MEMORY RateLimiterStoreType = "in_memory"
	RATE_LIMITER_STORE_REDIS                          = "redis"
)

func (r *RateLimiter) Validate() error {
	if r.Enabled {
		switch r.Store {
		case RATE_LIMITER_STORE_REDIS:
			if r.Redis == nil {
				return errors.New("when enabling the redis store you have to specify the redis config")
			}
			if r.Redis.Address == "" {
				return errors.New("when enabling the redis store you have to specify the address where hanko can reach the redis instance")
			}
		case RATE_LIMITER_STORE_IN_MEMORY:
			break
		default:
			return errors.New(string(r.Store) + " is not a valid rate limiter store.")
		}
	}
	return nil
}

type RedisConfig struct {
	// Address of redis in the form of host[:port][/database]
	Address  string `yaml:"address" json:"address" koanf:"address"`
	Password string `yaml:"password" json:"password,omitempty" koanf:"password"`
}

type ThirdParty struct {
	Providers             ThirdPartyProviders  `yaml:"providers" json:"providers,omitempty" koanf:"providers"`
	RedirectURL           string               `yaml:"redirect_url" json:"redirect_url,omitempty" koanf:"redirect_url" split_words:"true"`
	ErrorRedirectURL      string               `yaml:"error_redirect_url" json:"error_redirect_url,omitempty" koanf:"error_redirect_url" split_words:"true"`
	DefaultRedirectURL    string               `yaml:"default_redirect_url" json:"default_redirect_url,omitempty" koanf:"default_redirect_url" split_words:"true"`
	AllowedRedirectURLS   []string             `yaml:"allowed_redirect_urls" json:"allowed_redirect_urls,omitempty" koanf:"allowed_redirect_urls" split_words:"true"`
	AllowedRedirectURLMap map[string]glob.Glob `jsonschema:"-"`
}

func (t *ThirdParty) Validate() error {
	if t.Providers.HasEnabled() {
		if t.RedirectURL == "" {
			return errors.New("redirect_url must be set")
		}

		if t.ErrorRedirectURL == "" {
			return errors.New("error_redirect_url must be set")
		}

		if len(t.AllowedRedirectURLS) <= 0 {
			return errors.New("at least one allowed redirect url must be set")
		}

		urls := append(t.AllowedRedirectURLS, t.ErrorRedirectURL)
		if t.DefaultRedirectURL != "" {
			urls = append(urls, t.DefaultRedirectURL)
		}
		for _, u := range urls {
			if strings.HasSuffix(u, "/") {
				return fmt.Errorf("redirect url %s must not have trailing slash", u)
			}
		}
	}

	err := t.Providers.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate third party providers: %w", err)
	}

	return nil
}

func (t *ThirdParty) PostProcess() error {
	t.AllowedRedirectURLMap = make(map[string]glob.Glob)
	urls := append(t.AllowedRedirectURLS, t.ErrorRedirectURL)
	for _, redirectUrl := range urls {
		g, err := glob.Compile(redirectUrl, '.', '/')
		if err != nil {
			return fmt.Errorf("failed compile allowed redirect url glob: %w", err)
		}
		t.AllowedRedirectURLMap[redirectUrl] = g
	}

	return nil
}

type ThirdPartyProvider struct {
	Enabled      bool   `yaml:"enabled" json:"enabled" koanf:"enabled"`
	ClientID     string `yaml:"client_id" json:"client_id" koanf:"client_id" split_words:"true"`
	Secret       string `yaml:"secret" json:"secret" koanf:"secret"`
	AllowLinking bool   `yaml:"allow_linking" json:"allow_linking" koanf:"allow_linking" split_words:"true"`
	DisplayName  string `jsonschema:"-" yaml:"-" json:"-" koanf:"-"`
}

func (p *ThirdPartyProvider) Validate() error {
	if p.Enabled {
		if p.ClientID == "" {
			return errors.New("missing client ID")
		}
		if p.Secret == "" {
			return errors.New("missing client secret")
		}
	}
	return nil
}

type ThirdPartyProviders struct {
	Google    ThirdPartyProvider `yaml:"google" json:"google,omitempty" koanf:"google"`
	GitHub    ThirdPartyProvider `yaml:"github" json:"github,omitempty" koanf:"github"`
	Apple     ThirdPartyProvider `yaml:"apple" json:"apple,omitempty" koanf:"apple"`
	Discord   ThirdPartyProvider `yaml:"discord" json:"discord,omitempty" koanf:"discord"`
	Microsoft ThirdPartyProvider `yaml:"microsoft" json:"microsoft,omitempty" koanf:"microsoft"`
	LinkedIn  ThirdPartyProvider `yaml:"linkedin" json:"linkedin,omitempty" koanf:"linkedin"`
}

func (p *ThirdPartyProviders) Validate() error {
	s := structs.New(p)
	for _, field := range s.Fields() {
		provider := field.Value().(ThirdPartyProvider)
		err := provider.Validate()
		if err != nil {
			return fmt.Errorf("%s: %w", strings.ToLower(field.Name()), err)
		}
	}
	return nil
}

func (p *ThirdPartyProviders) HasEnabled() bool {
	s := structs.New(p)
	for _, field := range s.Fields() {
		provider := field.Value().(ThirdPartyProvider)
		if provider.Enabled {
			return true
		}
	}

	return false
}

func (p *ThirdPartyProviders) GetEnabled() []ThirdPartyProvider {
	s := structs.New(p)
	var enabledProviders []ThirdPartyProvider
	for _, field := range s.Fields() {
		provider := field.Value().(ThirdPartyProvider)
		if provider.Enabled {
			enabledProviders = append(enabledProviders, provider)
		}
	}

	return enabledProviders
}

func (p *ThirdPartyProviders) Get(provider string) *ThirdPartyProvider {
	s := structs.New(p)
	for _, field := range s.Fields() {
		if strings.ToLower(field.Name()) == strings.ToLower(provider) {
			p := field.Value().(ThirdPartyProvider)
			return &p
		}
	}

	return nil
}

func (c *Config) convertLegacyConfig() {
	c.Email.Limit = c.Emails.MaxNumOfAddresses
	c.Email.RequireVerification = c.Emails.RequireVerification
	c.Email.PasscodeTtl = c.Passcode.TTL

	c.EmailDelivery.SMTP = c.Smtp
	c.EmailDelivery.FromName = c.Passcode.Email.FromName
	c.EmailDelivery.FromAddress = c.Passcode.Email.FromAddress

	c.Passkey.UserVerification = c.Webauthn.UserVerification

	c.Webauthn.Timeouts.Login = c.Webauthn.Timeout
	c.Webauthn.Timeouts.Registration = c.Webauthn.Timeout
}

func (c *Config) PostProcess() error {
	c.arrangeSmtpSettings()

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

func (c *Config) arrangeSmtpSettings() {
	if !c.EmailDelivery.Enabled {
		return
	}
	if c.Passcode.Smtp.Validate() == nil {
		if c.Smtp.Validate() == nil {
			zeroLogger.Warn().Msg("Both root smtp and passcode.smtp are set. Using smtp settings from root configuration")
			return
		}

		c.Smtp = c.Passcode.Smtp
	}
}

type LoggerConfig struct {
	LogHealthAndMetrics bool `yaml:"log_health_and_metrics,omitempty" json:"log_health_and_metrics" koanf:"log_health_and_metrics" jsonschema:"default=true"`
}

type Account struct {
	// Allow Deletion indicates if a user can perform self-service deletion
	AllowDeletion bool `yaml:"allow_deletion" json:"allow_deletion,omitempty" koanf:"allow_deletion" jsonschema:"default=false"`
	AllowSignup   bool `yaml:"allow_signup" json:"allow_signup,omitempty" koanf:"allow_signup" jsonschema:"default=true"`
}

// TODO: below structs need validation, e.g. only allowed names for enabled and also we should reject some configurations (e.g. passcode & passwords are disabled and passkey onboarding is also disabled)

type SecondFactor struct {
	Enabled       bool                   `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	Optional      bool                   `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	Onboarding    SecondFactorOnboarding `yaml:"onboarding" json:"onboarding,omitempty" koanf:"onboarding"`
	Methods       []string               `yaml:"methods" json:"methods,omitempty" koanf:"methods" jsonschema:"enum=totp,enum=security_key"`
	RecoveryCodes RecoveryCodes          `yaml:"recovery_codes" json:"recovery_codes,omitempty" koanf:"recovery_codes" split_words:"true"`
}

type SecondFactorOnboarding struct {
	Enabled bool `yaml:"enabled" json:"enabled" koanf:"enabled"`
}

type RecoveryCodes struct {
	Enabled  bool `yaml:"enabled" json:"enabled" koanf:"enabled" jsonschema:"default=true"`
	Optional bool `yaml:"optional" json:"optional" koanf:"optional" jsonschema:"default=true"`
}

type Passkey struct {
	Enabled               bool   `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	Optional              bool   `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	AcquireOnRegistration string `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=always,enum=always,enum=conditional,enum=never"`
	AcquireOnLogin        string `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=always,enum=always,enum=conditional,enum=never"`
	UserVerification      string `yaml:"user_verification" json:"user_verification,omitempty" koanf:"user_verification" split_words:"true" jsonschema:"default=preferred,enum=required,enum=preferred,enum=discouraged"`
	AttestationPreference string `yaml:"attestation_preference" json:"attestation_preference,omitempty" koanf:"attestation_preference" split_words:"true" jsonschema:"default=direct,enum=direct,enum=indirect,enum=none"`
	Limit                 int    `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=100"`
}

type EmailDelivery struct {
	Enabled     bool   `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	FromAddress string `yaml:"from_address" json:"from_address,omitempty" koanf:"from_address" split_words:"true"`
	FromName    string `yaml:"from_name" json:"from_name,omitempty" koanf:"from_name" split_words:"true"`
	SMTP        SMTP   `yaml:"smtp" json:"smtp,omitempty" koanf:"smtp"`
}

type Email struct {
	Enabled               bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	Optional              bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=true"`
	AcquireOnLogin        bool `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=false"`
	RequireVerification   bool `yaml:"require_verification" json:"require_verification,omitempty" koanf:"require_verification" split_words:"true" jsonschema:"default=true"`
	Limit                 int  `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=100"`
	UseAsLoginIdentifier  bool `yaml:"use_as_login_identifier" json:"use_as_login_identifier,omitempty" koanf:"use_as_login_identifier" jsonschema:"default=true"`
	MaxLength             int  `yaml:"max_length" json:"max_length,omitempty" koanf:"max_length" jsonschema:"default=100"`
	UseForAuthentication  bool `yaml:"use_for_authentication" json:"use_for_authentication,omitempty" koanf:"use_for_authentication" jsonschema:"default=true"`
	PasscodeTtl           int  `yaml:"passcode_ttl" json:"passcode_ttl,omitempty" koanf:"passcode_ttl" jsonschema:"default=300"`
}

type Username struct {
	Enabled               bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	Optional              bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=false"`
	AcquireOnLogin        bool `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=false"`
	UseAsLoginIdentifier  bool `yaml:"use_as_login_identifier" json:"use_as_login_identifier,omitempty" koanf:"use_as_login_identifier" jsonschema:"default=true"`
	MinLength             int  `yaml:"min_length" json:"min_length,omitempty" koanf:"min_length" split_words:"true" jsonschema:"default=8"`
	MaxLength             int  `yaml:"max_length" json:"max_length,omitempty" koanf:"max_length" jsonschema:"default=100"`
}
