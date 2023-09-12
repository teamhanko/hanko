package config

import (
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/gobwas/glob"
	"github.com/kelseyhightower/envconfig"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"golang.org/x/exp/slices"
	"log"
	"strings"
	"time"
)

// Config is the central configuration type
type Config struct {
	Server      Server           `yaml:"server" json:"server,omitempty" koanf:"server"`
	Webauthn    WebauthnSettings `yaml:"webauthn" json:"webauthn,omitempty" koanf:"webauthn"`
	Passcode    Passcode         `yaml:"passcode" json:"passcode" koanf:"passcode"`
	Password    Password         `yaml:"password" json:"password,omitempty" koanf:"password"`
	Database    Database         `yaml:"database" json:"database" koanf:"database"`
	Secrets     Secrets          `yaml:"secrets" json:"secrets" koanf:"secrets"`
	Service     Service          `yaml:"service" json:"service" koanf:"service"`
	Session     Session          `yaml:"session" json:"session,omitempty" koanf:"session"`
	AuditLog    AuditLog         `yaml:"audit_log" json:"audit_log,omitempty" koanf:"audit_log" split_words:"true"`
	Emails      Emails           `yaml:"emails" json:"emails,omitempty" koanf:"emails"`
	RateLimiter RateLimiter      `yaml:"rate_limiter" json:"rate_limiter,omitempty" koanf:"rate_limiter" split_words:"true"`
	ThirdParty  ThirdParty       `yaml:"third_party" json:"third_party,omitempty" koanf:"third_party" split_words:"true"`
	Log         LoggerConfig     `yaml:"log" json:"log,omitempty" koanf:"log"`
	Account     Account          `yaml:"account" json:"account,omitempty" koanf:"account"`
}

var (
	DefaultConfigFilePath = "./config/config.yaml"
)

func Load(cfgFile *string) (*Config, error) {
	k := koanf.New(".")
	var err error
	if cfgFile == nil || *cfgFile == "" {
		*cfgFile = "./config/config.yaml"
	}
	if err = k.Load(file.Provider(*cfgFile), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load config from: %s: %w", *cfgFile, err)
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

func DefaultConfig() *Config {
	return &Config{
		Server: Server{
			Public: ServerSettings{
				Address: ":8000",
			},
			Admin: ServerSettings{
				Address: ":8001",
			},
		},
		Webauthn: WebauthnSettings{
			RelyingParty: RelyingParty{
				Id:          "localhost",
				DisplayName: "Hanko Authentication Service",
				Origins:     []string{"http://localhost:8888"},
			},
			UserVerification: "preferred",
			Timeout:          60000,
		},
		Passcode: Passcode{
			Smtp: SMTP{
				Port: "465",
			},
			TTL: 300,
			Email: Email{
				FromAddress: "passcode@hanko.io",
				FromName:    "Hanko",
			},
		},
		Password: Password{
			MinPasswordLength: 8,
		},
		Database: Database{
			Database: "hanko",
		},
		Session: Session{
			Lifespan: "1h",
			Cookie: Cookie{
				HttpOnly: true,
				SameSite: "strict",
				Secure:   true,
			},
		},
		AuditLog: AuditLog{
			ConsoleOutput: AuditLogConsole{
				Enabled:      true,
				OutputStream: OutputStreamStdOut,
			},
		},
		Emails: Emails{
			RequireVerification: true,
			MaxNumOfAddresses:   5,
		},
		RateLimiter: RateLimiter{
			Enabled: true,
			Store:   RATE_LIMITER_STORE_IN_MEMORY,
			PasswordLimits: RateLimits{
				Tokens:   5,
				Interval: 1 * time.Minute,
			},
			PasscodeLimits: RateLimits{
				Tokens:   3,
				Interval: 1 * time.Minute,
			},
			TokenLimits: RateLimits{
				Tokens:   3,
				Interval: 1 * time.Minute,
			},
		},
		Account: Account{
			AllowDeletion: false,
			AllowSignup:   true,
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
	Name string `yaml:"name" json:"name" koanf:"name"`
}

func (s *Service) Validate() error {
	if len(strings.TrimSpace(s.Name)) == 0 {
		return errors.New("field name must not be empty")
	}
	return nil
}

type Password struct {
	Enabled           bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	MinPasswordLength int  `yaml:"min_password_length" json:"min_password_length,omitempty" koanf:"min_password_length" split_words:"true" jsonschema:"default=8"`
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
	Address    string `yaml:"address" json:"address,omitempty" koanf:"address"`
	PathPrefix string `yaml:"path_prefix" json:"path_prefix,omitempty" koanf:"path_prefix" split_words:"true"`
	Cors       Cors   `yaml:"cors" json:"cors,omitempty" koanf:"cors"`
}

type Cors struct {
	// AllowOrigins determines the value of the Access-Control-Allow-Origin
	// response header. This header defines a list of origins that may access the
	// resource.  The wildcard characters '*' and '?' are supported and are
	// converted to regex fragments '.*' and '.' accordingly.
	AllowOrigins []string `yaml:"allow_origins" json:"allow_origins" koanf:"allow_origins" split_words:"true"`

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

// WebauthnSettings defines the settings for the webauthn authentication mechanism
type WebauthnSettings struct {
	RelyingParty     RelyingParty `yaml:"relying_party" json:"relying_party,omitempty" koanf:"relying_party" split_words:"true"`
	Timeout          int          `yaml:"timeout" json:"timeout,omitempty" koanf:"timeout" jsonschema:"default=60000"`
	UserVerification string       `yaml:"user_verification" json:"user_verification,omitempty" koanf:"user_verification" split_words:"true" jsonschema:"default=preferred,enum=required,enum=preferred,enum=discouraged"`
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
	Host     string `yaml:"host" json:"host" koanf:"host"`
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

type Email struct {
	FromAddress string `yaml:"from_address" json:"from_address,omitempty" koanf:"from_address" split_words:"true" jsonschema:"default=passcode@hanko.io"`
	FromName    string `yaml:"from_name" json:"from_name,omitempty" koanf:"from_name" split_words:"true" jsonschema:"default=Hanko"`
}

func (e *Email) Validate() error {
	if len(strings.TrimSpace(e.FromAddress)) == 0 {
		return errors.New("from_address must not be empty")
	}
	return nil
}

type Passcode struct {
	Email Email `yaml:"email" json:"email,omitempty" koanf:"email"`
	Smtp  SMTP  `yaml:"smtp" json:"smtp" koanf:"smtp"`
	TTL   int   `yaml:"ttl" json:"ttl,omitempty" koanf:"ttl" jsonschema:"default=300"`
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
	Database string `yaml:"database" json:"database,omitempty" koanf:"database" jsonschema:"default=hanko" jsonschema:"oneof_required=config"`
	User     string `yaml:"user" json:"user,omitempty" koanf:"user" jsonschema:"oneof_required=config"`
	Password string `yaml:"password" json:"password,omitempty" koanf:"password" jsonschema:"oneof_required=config"`
	Host     string `yaml:"host" json:"host,omitempty" koanf:"host" jsonschema:"oneof_required=config"`
	Port     string `yaml:"port" json:"port,omitempty" koanf:"port" jsonschema:"oneof_required=config,oneof_type=string;integer"`
	Dialect  string `yaml:"dialect" json:"dialect,omitempty" koanf:"dialect" jsonschema:"oneof_required=config,enum=postgres,enum=mysql,enum=cockroach"`
	Url      string `yaml:"url" json:"url,omitempty" koanf:"url" jsonschema:"oneof_required=url"`
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
	Keys []string `yaml:"keys" json:"keys" koanf:"keys" jsonschema:"minItems=1"`
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

	// EnableRefreshToken optional bool to enable refresh tokens. If set to true, refresh tokens will be issued.
	EnableRefreshToken bool `yaml:"enable_refresh_token" json:"enable_refresh_token,omitempty" koanf:"enable_refresh_token" split_words:"true" jsonschema:"default=false"`

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
	for _, url := range urls {
		g, err := glob.Compile(url, '.', '/')
		if err != nil {
			return fmt.Errorf("failed compile allowed redirect url glob: %w", err)
		}
		t.AllowedRedirectURLMap[url] = g
	}

	return nil
}

type ThirdPartyProvider struct {
	Enabled  bool   `yaml:"enabled" json:"enabled" koanf:"enabled"`
	ClientID string `yaml:"client_id" json:"client_id" koanf:"client_id" split_words:"true"`
	Secret   string `yaml:"secret" json:"secret" koanf:"secret"`
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
	Google ThirdPartyProvider `yaml:"google" json:"google,omitempty" koanf:"google"`
	GitHub ThirdPartyProvider `yaml:"github" json:"github,omitempty" koanf:"github"`
	Apple  ThirdPartyProvider `yaml:"apple" json:"apple,omitempty" koanf:"apple"`
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

func (c *Config) PostProcess() error {
	err := c.ThirdParty.PostProcess()
	if err != nil {
		return fmt.Errorf("failed to post process third party settings: %w", err)
	}

	return nil

}

type LoggerConfig struct {
	LogHealthAndMetrics bool `yaml:"log_health_and_metrics,omitempty" json:"log_health_and_metrics" koanf:"log_health_and_metrics" jsonschema:"default=true"`
}

type Account struct {
	// Allow Deletion indicates if a user can perform self-service deletion
	AllowDeletion bool `yaml:"allow_deletion" json:"allow_deletion,omitempty" koanf:"allow_deletion" jsonschema:"default=false"`
	AllowSignup   bool `yaml:"allow_signup" json:"allow_signup,omitempty" koanf:"allow_signup" jsonschema:"default=true"`
}
