package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/fatih/structs"
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobwas/glob"
	"github.com/kelseyhightower/envconfig"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/teamhanko/hanko/backend/ee/saml/config"
	"golang.org/x/exp/slices"
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
	// `passkey` configures how passkeys are acquired and used.
	Passkey Passkey `yaml:"passkey" json:"passkey,omitempty" koanf:"passkey" jsonschema:"title=passkey"`
	// `passlink` congigures how passlinks are acquired and used.
	Passlink Passlink `yaml:"passlink" json:"passlink,omitempty" koanf:"passlink"`
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

type Server struct {
	// `public` contains the server configuration for the public API.
	Public ServerSettings `yaml:"public" json:"public,omitempty" koanf:"public" jsonschema:"title=public"`
	// `admin` contains the server configuration for the admin API.
	Admin ServerSettings `yaml:"admin" json:"admin,omitempty" koanf:"admin" jsonschema:"title=admin"`
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
	// `name` determines the name of the service.
	// This value is used, e.g. in the subject header of outgoing emails.
	Name string `yaml:"name" json:"name,omitempty" koanf:"name"`
}

func (s *Service) Validate() error {
	if len(strings.TrimSpace(s.Name)) == 0 {
		return errors.New("field name must not be empty")
	}
	return nil
}

type Password struct {
	// `acquire_on_registration` configures how users are prompted creating a password on registration.
	AcquireOnRegistration string `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=always,enum=always,enum=conditional,enum=never"`
	// `acquire_on_login` configures how users are prompted creating a password on login.
	AcquireOnLogin string `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=never,enum=always,enum=conditional,enum=never"`
	// `enabled` determines whether passwords are enabled or disabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `min_length` determines the minimum password length.
	MinLength int `yaml:"min_length" json:"min_length,omitempty" koanf:"min_length" split_words:"true" jsonschema:"default=8"`
	// Deprecated. Use `min_length` instead.
	MinPasswordLength int `yaml:"min_password_length" json:"min_password_length,omitempty" koanf:"min_password_length" split_words:"true" jsonschema:"default=8"`
	// `optional` determines whether users must set a password when prompted. The password cannot be deleted if
	// passwords are required (`optional: false`).
	//
	// It also takes part in determining the order of password and passkey acquisition
	// on login and registration (see also `acquire_on_login` and `acquire_on_registration`): if one credential type is
	// required (`optional: false`) then that one takes precedence, i.e. is acquired first.
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=false"`
	// `recovery` determines whether users can start a recovery process, e.g. in case of a forgotten password.
	Recovery bool `yaml:"recovery" json:"recovery,omitempty" koanf:"recovery" jsonschema:"default=true"`
}

func (Password) JSONSchemaExtend(schema *jsonschema.Schema) {
	acquireOnRegistration, _ := schema.Properties.Get("acquire_on_registration")
	acquireOnRegistration.Extras = map[string]any{"meta:enum": map[string]string{
		"always": "Indicates that users are always prompted to create a password on registration.",
		"conditional": `Indicates that users are prompted to create a password on registration as long as the user does
						not have a passkey.

						If passkeys are also conditionally acquired on registration, then users are given a choice as
						to what type of credential to register.`,
		"never": "Indicates that users are never prompted to create a password on registration.",
	}}

	acquireOnLogin, _ := schema.Properties.Get("acquire_on_login")
	acquireOnLogin.Extras = map[string]any{"meta:enum": map[string]string{
		"always": `Indicates that users are always prompted to create a password on login
					provided that they do not already have a password.`,
		"conditional": `Indicates that users are prompted to create a password on login provided that
						they do not already have a password and do not have a passkey.

						If passkeys are also conditionally acquired on login then users are given a choice as to what
						type of credential to register.`,
		"never": "Indicates that users are never prompted to create a password on login.",
	}}
}

type Cookie struct {
	// `domain` is the domain the cookie will be bound to. Works for subdomains, but not cross-domain.
	// See the `session.enable_auth_token_header` configuration instead if the API and the client application run on
	// different domains.
	Domain string `yaml:"domain" json:"domain,omitempty" koanf:"domain" jsonschema:"default=hanko"`
	// `http_only` determines whether cookies are HTTP only or accessible by Javascript.
	HttpOnly bool `yaml:"http_only" json:"http_only,omitempty" koanf:"http_only" split_words:"true" jsonschema:"default=true"`
	// `name` is the name of the cookie.
	Name string `yaml:"name" json:"name,omitempty" koanf:"name" jsonschema:"default=hanko"`
	// `same_site` controls whether a cookie is sent with cross-site requests.
	// See [here](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#samesitesamesite-value) for
	// more details.
	SameSite string `yaml:"same_site" json:"same_site,omitempty" koanf:"same_site" split_words:"true" jsonschema:"default=strict,enum=strict,enum=lax,enum=none"`
	// `secure` indicates whether the cookie is sent to the server only when a request is made with the https: scheme
	// (except on localhost).
	//
	// NOTE: `secure` must be set to `false` when working on `localhost` and with the Safari browser because it does
	// not store secure cookies on `localhost`.
	Secure bool `yaml:"secure" json:"secure,omitempty" koanf:"secure" jsonschema:"default=true"`
}

func (c *Cookie) GetName() string {
	if c.Name != "" {
		return c.Name
	}

	return "hanko"
}

type ServerSettings struct {
	// `address` is the address of the server to listen on in the form of host:port.
	//
	// See [net.Dial](https://pkg.go.dev/net#Dial) for details of the address format.
	Address string `yaml:"address" json:"address,omitempty" koanf:"address"`
	// `cors` contains configuration options regarding Cross-Origin-Resource-Sharing.
	Cors Cors `yaml:"cors" json:"cors,omitempty" koanf:"cors" jsonschema:"title=cors"`
}

type Cors struct {
	// `allow_origins` determines the value of the Access-Control-Allow-Origin
	// response header. This header defines a list of [origins](https://developer.mozilla.org/en-US/docs/Glossary/Origin)
	// that may access the resource.
	//
	// The wildcard characters `*` and `?` are supported and are converted to regex fragments `.*` and `.` accordingly.
	AllowOrigins []string `yaml:"allow_origins" json:"allow_origins,omitempty" koanf:"allow_origins" split_words:"true" jsonschema:"title=allow_origins,default=http://localhost:8888"`

	// `unsafe_wildcard_origin_allowed` allows a wildcard `*` origin to be used with AllowCredentials
	// flag. In that case we consider any origin allowed and send it back to the client in an `Access-Control-Allow-Origin` header.
	//
	// This is INSECURE and potentially leads to [cross-origin](https://portswigger.net/research/exploiting-cors-misconfigurations-for-bitcoins-and-bounties)
	// attacks. See also https://github.com/labstack/echo/issues/2400 for discussion on the subject.
	//
	// Optional. Default value is `false`.
	UnsafeWildcardOriginAllowed bool `yaml:"unsafe_wildcard_origin_allowed" json:"unsafe_wildcard_origin_allowed,omitempty" koanf:"unsafe_wildcard_origin_allowed" split_words:"true" jsonschema:"title=unsafe_wildcard_origin_allowed,default=false"`
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
	// `registration` determines the time, in milliseconds, that the client is willing to wait for the credential
	// creation request to the WebAuthn API to complete.
	Registration int `yaml:"registration" json:"registration,omitempty" koanf:"registration" jsonschema:"default=600000"`
	// `login` determines the time, in milliseconds, that the client is willing to wait for the credential
	//  request to the WebAuthn API to complete.
	Login int `yaml:"login" json:"login,omitempty" koanf:"login" jsonschema:"default=600000"`
}

// WebauthnSettings defines the settings for the webauthn authentication mechanism
type WebauthnSettings struct {
	RelyingParty RelyingParty `yaml:"relying_party" json:"relying_party,omitempty" koanf:"relying_party" split_words:"true" jsonschema:"title=relying_party"`
	// Deprecated, use `timeouts` instead.
	Timeout int `yaml:"timeout" json:"timeout,omitempty" koanf:"timeout" jsonschema:"default=60000"`
	// `timeouts` specifies the timeouts for passkey/WebAuthn registration and login.
	Timeouts WebauthnTimeouts `yaml:"timeouts" json:"timeouts,omitempty" koanf:"timeouts" split_words:"true" jsonschema:"title=timeouts"`
	// Deprecated, use `passkey.user_verification` instead
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
	// `display_name` is the service's name that some WebAuthn Authenticators will display to the user during registration
	// and authentication ceremonies.
	DisplayName string `yaml:"display_name" json:"display_name,omitempty" koanf:"display_name" split_words:"true" jsonschema:"default=Hanko Authentication Service"`
	Icon        string `yaml:"icon" json:"icon,omitempty" koanf:"icon" jsonschema:"-"`
	// `id` is the [effective domain](https://html.spec.whatwg.org/multipage/browsers.html#concept-origin-effective-domain)
	// the passkey/WebAuthn credentials will be bound to.
	Id string `yaml:"id" json:"id,omitempty" koanf:"id" jsonschema:"default=localhost,examples=localhost,example.com,subdomain.example.com"`
	// `origins` is a list of origins for which passkeys/WebAuthn credentials will be accepted by the server. Must
	// include the protocol and can only be the effective domain, or a registrable domain suffix of the effective
	// domain, as specified in the [`id`](#id). Except for `localhost`, the protocol **must** always be `https` for
	// passkeys/WebAuthn to work. IP Addresses will not work.
	//
	// For an Android application the origin must be the base64 url encoded SHA256 fingerprint of the signing
	// certificate.
	Origins []string `yaml:"origins" json:"origins,omitempty" koanf:"origins" jsonschema:"minItems=1,default=http://localhost:8888,examples=android:apk-key-hash:nLSu7wVTbnMOxLgC52f2faTnvCbXQrUn_wF9aCrr-l0,https://login.example.com"`
}

// SMTP Server Settings for sending passcodes
type SMTP struct {
	Host     string `yaml:"host" json:"host,omitempty" koanf:"host" jsonschema:"default=localhost"`
	Port     string `yaml:"port" json:"port,omitempty" koanf:"port" jsonschema:"default=465"`
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

type Passcode struct {
	// Deprecated. Use `email.passcode_ttl` instead.
	TTL int `yaml:"ttl" json:"ttl,omitempty" koanf:"ttl" jsonschema:"default=300"`
}

type Database struct {
	// `database` determines the name of the database schema to use.
	Database string `yaml:"database" json:"database,omitempty" koanf:"database" jsonschema:"default=hanko"`
	// `dialect` is the name of the database system to use.
	Dialect string `yaml:"dialect" json:"dialect,omitempty" koanf:"dialect" jsonschema:"default=postgres,enum=postgres,enum=mysql,enum=mariadb,enum=cockroach"`
	// `host` is the host the database system is running on.
	Host string `yaml:"host" json:"host,omitempty" koanf:"host" jsonschema:"default=localhost"`
	// `password` is the password for the database user to use for connecting to the database.
	Password string `yaml:"password" json:"password,omitempty" koanf:"password" jsonschema:"default=hanko"`
	// `port` is the port the database system is running on.
	Port string `yaml:"port" json:"port,omitempty" koanf:"port" jsonschema:"default=5432"`
	// `url` is a datasource connection string. It can be used instead of the rest of the database configuration
	// options. If this `url` is set then it is prioritized, i.e. the rest of the options, if set, have no effect.
	//
	// Schema: `dialect://username:password@host:port/database`
	Url string `yaml:"url" json:"url,omitempty" koanf:"url" jsonschema:"example=postgres://hanko:hanko@localhost:5432/hanko"`
	// `user` is the database user to use for connecting to the database.
	User string `yaml:"user" json:"user,omitempty" koanf:"user" jsonschema:"default=hanko"`
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
	// `keys` are used to en- and decrypt the JWKs which get used to sign the JWTs issued by the API.
	// For every key a JWK is generated, encrypted with the key and persisted in the database.
	//
	// You can use this list for key rotation: add a new key to the beginning of the list and the corresponding
	// JWK will then be used for signing JWTs. All tokens signed with the previous JWK(s) will still
	// be valid until they expire. Removing a key from the list does not remove the corresponding
	// database record. If you remove a key, you also have to remove the database record, otherwise
	// application startup will fail.
	Keys []string `yaml:"keys" json:"keys,omitempty" koanf:"keys" jsonschema:"minItems=1"`
}

func (Secrets) JSONSchemaExtend(schema *jsonschema.Schema) {
	keys, _ := schema.Properties.Get("keys")
	var keysItemsMinLength uint64 = 16
	keys.Items = &jsonschema.Schema{
		Type:      "string",
		Title:     "keys",
		MinLength: &keysItemsMinLength,
	}
}

func (s *Secrets) Validate() error {
	if len(s.Keys) == 0 {
		return errors.New("at least one key must be defined")
	}
	return nil
}

type Session struct {
	// `audience` is a list of strings that identifies the recipients that the JWT is intended for.
	// The audiences are placed in the `aud` claim of the JWT.
	// If not set, it defaults to the value of the`webauthn.relying_party.id` configuration parameter.
	Audience []string `yaml:"audience" json:"audience,omitempty" koanf:"audience"`
	// `cookie` contains configuration for the session cookie issued on successful registration or login.
	Cookie Cookie `yaml:"cookie" json:"cookie,omitempty" koanf:"cookie"`
	// `enable_auth_token_header` determines whether a session token (JWT) is returned in an `X-Auth-Token`
	// header after a successful authentication. This option should be set to `true` if API and client applications
	// run on different domains.
	EnableAuthTokenHeader bool `yaml:"enable_auth_token_header" json:"enable_auth_token_header,omitempty" koanf:"enable_auth_token_header" split_words:"true" jsonschema:"default=false"`
	// `issuer` is a string that identifies the principal (human user, an organization, or a service)
	// that issued the JWT. Its value is set in the `iss` claim of a JWT.
	Issuer string `yaml:"issuer" json:"issuer,omitempty" koanf:"issuer"`
	// `lifespan` determines how long a session token (JWT) is valid. It must be a (possibly signed) sequence of decimal
	// numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Lifespan string `yaml:"lifespan" json:"lifespan,omitempty" koanf:"lifespan" jsonschema:"default=12h"`
}

func (s *Session) Validate() error {
	_, err := time.ParseDuration(s.Lifespan)
	if err != nil {
		return errors.New("failed to parse lifespan")
	}
	return nil
}

type AuditLog struct {
	// `console_output` controls audit log console output.
	ConsoleOutput AuditLogConsole `yaml:"console_output" json:"console_output,omitempty" koanf:"console_output" split_words:"true" jsonschema:"title=console_output"`
	// `mask` determines whether sensitive information (usernames, emails) should be masked in the audit log output.
	//
	// This configuration applies to logs written to the console as well as persisted logs.
	Mask bool `yaml:"mask" json:"mask,omitempty" koanf:"mask" jsonschema:"default=true"`
	// `storage` controls audit log retention.
	Storage AuditLogStorage `yaml:"storage" json:"storage,omitempty" koanf:"storage"`
}

type AuditLogStorage struct {
	// `enabled` controls whether audit log should be retained (i.e. persisted).
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
}

type AuditLogConsole struct {
	// `enabled` controls whether audit log output on the console is enabled or disabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `output` determines the output stream audit logs are sent to.
	OutputStream OutputStream `yaml:"output" json:"output,omitempty" koanf:"output" split_words:"true" jsonschema:"default=stdout,enum=stdout,enum=stderr"`
}

type Emails struct {
	// Deprecated. Use `email.require_verification` instead.
	RequireVerification bool `yaml:"require_verification" json:"require_verification,omitempty" koanf:"require_verification" split_words:"true" jsonschema:"default=true"`
	// Deprecated. Use `email.limit` instead.
	MaxNumOfAddresses int `yaml:"max_num_of_addresses" json:"max_num_of_addresses,omitempty" koanf:"max_num_of_addresses" split_words:"true" jsonschema:"default=5"`
}

type OutputStream string

var (
	OutputStreamStdOut OutputStream = "stdout"
	OutputStreamStdErr OutputStream = "stderr"
)

type RateLimiter struct {
	// `enabled` controls whether rate limiting is enabled or disabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `store` sets the store for the rate limiter. When you have multiple instances of Hanko running, it is recommended to use
	//  the `redis` store because otherwise your instances each have their own states.
	Store RateLimiterStoreType `yaml:"store" json:"store,omitempty" koanf:"store" jsonschema:"default=in_memory,enum=in_memory,enum=redis"`
	// `redis_config` configures connection to a redis instance.
	// Required if `store` is set to `redis`
	Redis *RedisConfig `yaml:"redis_config" json:"redis_config,omitempty" koanf:"redis_config"`
	// `passcode_limits` controls rate limits for passcode operations.
	PasscodeLimits RateLimits `yaml:"passcode_limits" json:"passcode_limits,omitempty" koanf:"passcode_limits" split_words:"true"`
	// `passlink_limits` controls rate limits for passlink operations.
	PasslinkLimits RateLimits `yaml:"passlink_limits" json:"passlink_limits,omitempty" koanf:"passlink_limits" split_words:"true"`
	// `password_limits` controls rate limits for password login operations.
	PasswordLimits RateLimits `yaml:"password_limits" json:"password_limits,omitempty" koanf:"password_limits" split_words:"true"`
	// `token_limits` controls rate limits for token exchange operations.
	TokenLimits RateLimits `yaml:"token_limits" json:"token_limits,omitempty" koanf:"token_limits" split_words:"true" jsonschema:"default=token=3;interval=1m"`
}

type RateLimits struct {
	// `tokens` determines how many operations/requests can occur in the given `interval`.
	Tokens uint64 `yaml:"tokens" json:"tokens" koanf:"tokens" jsonschema:"default=3"`
	// `interval` determines when to reset the token interval.
	// It must be a (possibly signed) sequence of decimal
	// numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Interval time.Duration `yaml:"interval" json:"interval" koanf:"interval" jsonschema:"default=1m,type=string"`
}

type RateLimiterStoreType string

const (
	RATE_LIMITER_STORE_IN_MEMORY RateLimiterStoreType = "in_memory"
	RATE_LIMITER_STORE_REDIS     RateLimiterStoreType = "redis"
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
	// `address` is the address of the redis instance in the form of `host[:port][/database]`.
	Address string `yaml:"address" json:"address" koanf:"address"`
	// `password` is the password for the redis instance.
	Password string `yaml:"password" json:"password,omitempty" koanf:"password"`
}

type ThirdParty struct {
	// `providers` contains the configurations for the available OAuth/OIDC identity providers.
	Providers ThirdPartyProviders `yaml:"providers" json:"providers,omitempty" koanf:"providers" jsonschema:"title=providers,uniqueItems=true"`
	// `redirect_url` is the URL the third party provider redirects to with an authorization code. Must consist of the base URL
	// of your running Hanko backend instance and the `callback` endpoint of the API,
	// i.e. `{YOUR_BACKEND_INSTANCE}/thirdparty/callback.`
	//
	// Required if any of the [`providers`](#providers) are `enabled`.
	RedirectURL string `yaml:"redirect_url" json:"redirect_url,omitempty" koanf:"redirect_url" split_words:"true" jsonschema:"example=https://yourinstance.com/thirdparty/callback"`
	// `error_redirect_url` is the URL the backend redirects to if an error occurs during third party sign-in.
	// Errors are provided as 'error' and 'error_description' query params in the redirect location URL.
	//
	// When using the Hanko web components it should be the URL of the page that embeds the web component such that
	// errors can be processed properly by the web component.
	//
	// You do not have to add this URL to the 'allowed_redirect_urls', it is automatically included when validating
	// redirect URLs.
	//
	// Required if any of the [`providers`](#providers) are `enabled`. Must not have trailing slash.
	ErrorRedirectURL string `yaml:"error_redirect_url" json:"error_redirect_url,omitempty" koanf:"error_redirect_url" split_words:"true"`
	// `default_redirect_url` is the URL the backend redirects to after it successfully verified
	// the response from any third party provider.
	//
	// Must not have trailing slash.
	DefaultRedirectURL string `yaml:"default_redirect_url" json:"default_redirect_url,omitempty" koanf:"default_redirect_url" split_words:"true"`
	// `allowed_redirect_urls` is a list of URLs the backend is allowed to redirect to after third party sign-in was
	// successful.
	//
	// Supports wildcard matching through globbing. e.g. `https://*.example.com` will allow `https://foo.example.com`
	// and `https://bar.example.com` to be accepted.
	//
	// Globbing is also supported for paths, e.g. `https://foo.example.com/*` will match `https://foo.example.com/page1`
	// and `https://foo.example.com/page2`.
	//
	// A double asterisk (`**`) acts as a "super"-wildcard/match-all.
	//
	// See [here](https://pkg.go.dev/github.com/gobwas/glob#Compile) for more on globbing.
	//
	// Must not be empty if any of the [`providers`](#providers) are `enabled`. URLs in the list must not have a trailing slash.
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
	// `allow_linking` indicates whether existing accounts can be automatically linked with this provider.
	//
	// Linking is based on matching one of the email addresses of an existing user account with the (primary)
	// email address of the third party provider account.
	AllowLinking bool `yaml:"allow_linking" json:"allow_linking,omitempty" koanf:"allow_linking" split_words:"true"`
	// `client_id` is the ID of the OAuth/OIDC client. Must be obtained from the provider.
	//
	// Required if the provider is `enabled`.
	ClientID    string `yaml:"client_id" json:"client_id,omitempty" koanf:"client_id" split_words:"true"`
	DisplayName string `jsonschema:"-" yaml:"-" json:"-" koanf:"-"`
	// `enabled` determines whether this provider is enabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `secret` is the client secret for the OAuth/OIDC client. Must be obtained from the provider.
	//
	// Required if the provider is `enabled`.
	Secret string `yaml:"secret" json:"secret,omitempty" koanf:"secret"`
}

func (ThirdPartyProvider) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Title = "provider"

	enabledTrue := &jsonschema.Schema{Properties: orderedmap.New[string, *jsonschema.Schema]()}
	enabledTrue.Properties.Set("enabled", &jsonschema.Schema{Const: true})

	schema.If = enabledTrue
	schema.Then = &jsonschema.Schema{
		Required: []string{"client_id", "secret"},
	}
	schema.Else = &jsonschema.Schema{
		Required: []string{"enabled"},
	}
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
	// `apple` contains the provider configuration for Apple.
	Apple ThirdPartyProvider `yaml:"apple" json:"apple,omitempty" koanf:"apple"`
	// `discord` contains the provider configuration for Discord.
	Discord ThirdPartyProvider `yaml:"discord" json:"discord,omitempty" koanf:"discord"`
	// `github` contains the provider configuration for GitHub.
	GitHub ThirdPartyProvider `yaml:"github" json:"github,omitempty" koanf:"github"`
	// `google` contains the provider configuration for Google.
	Google ThirdPartyProvider `yaml:"google" json:"google,omitempty" koanf:"google"`
	// `linkedin` contains the provider configuration for LinkedIn.
	LinkedIn ThirdPartyProvider `yaml:"linkedin" json:"linkedin,omitempty" koanf:"linkedin"`
	// `microsoft` contains the provider configuration for Microsoft.
	Microsoft ThirdPartyProvider `yaml:"microsoft" json:"microsoft,omitempty" koanf:"microsoft"`
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
		if strings.EqualFold(field.Name(), provider) {
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

type LoggerConfig struct {
	// `log_health_and_metrics` determines whether requests of the `/health` and `/metrics` endpoints are logged.
	LogHealthAndMetrics bool `yaml:"log_health_and_metrics,omitempty" json:"log_health_and_metrics" koanf:"log_health_and_metrics" jsonschema:"default=true"`
}

type Account struct {
	// `allow_deletion` determines whether users can delete their accounts.
	AllowDeletion bool `yaml:"allow_deletion" json:"allow_deletion,omitempty" koanf:"allow_deletion" jsonschema:"default=false"`
	// `allow_signup` determines whether users are able to create new accounts.
	AllowSignup bool `yaml:"allow_signup" json:"allow_signup,omitempty" koanf:"allow_signup" jsonschema:"default=true"`
}

type Passkey struct {
	// `acquire_on_registration` configures how users are prompted creating a passkey on registration.
	AcquireOnRegistration string `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=always,enum=always,enum=conditional,enum=never"`
	// `acquire_on_login` configures how users are prompted creating a passkey on login.
	AcquireOnLogin string `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=always,enum=always,enum=conditional,enum=never"`
	// `attestation_preference` is used to specify the preference regarding attestation conveyance during
	// credential generation.
	AttestationPreference string `yaml:"attestation_preference" json:"attestation_preference,omitempty" koanf:"attestation_preference" split_words:"true" jsonschema:"default=direct,enum=direct,enum=indirect,enum=none"`
	// `enabled` determines whether users can create or authenticate with passkeys.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `limit` defines the maximum number of passkeys a user can have.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=10"`
	// `optional` determines whether users must create a passkey when prompted. The last remaining passkey cannot be
	// deleted if passkeys are required (`optional: false`).
	//
	// It also takes part in determining the order of password and passkey acquisition
	// on login and registration (see also `acquire_on_login` and `acquire_on_registration`): if one credential type is
	// required (`optional: false`) then that one takes precedence, i.e. is acquired first.
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	// `user_verification` specifies the requirements regarding local authorization with an authenticator through
	//  various authorization gesture modalities; for example, through a touch plus pin code,
	//  password entry, or biometric recognition.
	//
	// The setting applies to both WebAuthn registration and authentication ceremonies.
	UserVerification string `yaml:"user_verification" json:"user_verification,omitempty" koanf:"user_verification" split_words:"true" jsonschema:"default=preferred,enum=required,enum=preferred,enum=discouraged"`
}

func (Passkey) JSONSchemaExtend(schema *jsonschema.Schema) {
	acquireOnRegistration, _ := schema.Properties.Get("acquire_on_registration")
	acquireOnRegistration.Extras = map[string]any{"meta:enum": map[string]string{
		"always": "Indicates that users are always prompted to create a passkey on registration.",
		"conditional": `Indicates that users are prompted to create a passkey on registration as long as the user does
						not have a password.

						If passwords are also conditionally acquired on registration, then users are given a choice as
						to what type of credential to create.`,
		"never": "Indicates that users are never prompted to create a passkey on registration.",
	}}

	acquireOnLogin, _ := schema.Properties.Get("acquire_on_login")
	acquireOnLogin.Extras = map[string]any{"meta:enum": map[string]string{
		"always": `Indicates that users are always prompted to create a passkey on login
					provided that they do not already have a passkey.`,
		"conditional": `Indicates that users are prompted to create a passkey on login provided that
						they do not already have a passkey and do not have a password.

						If passkeys are also conditionally acquired on login then users are given a choice as to what
						type of credential to register.`,
		"never": "Indicates that users are never prompted to create a passkey on login.",
	}}

	userVerification, _ := schema.Properties.Get("user_verification")
	userVerification.Extras = map[string]any{"meta:enum": map[string]string{
		"required": "Indicates that user verification is always required.",
		"preferred": `Indicates that user verification is preferred but will not fail the operation if no
						user verification was performed.`,
		"discouraged": "Indicates that no user verification should be performed.",
	}}

	attestationPreference, _ := schema.Properties.Get("attestation_preference")
	attestationPreference.Extras = map[string]any{"meta:enum": map[string]string{
		"direct": `Indicates that the Relying Party wants to receive the attestation statement as generated by
					the authenticator.`,
		"indirect": `Indicates that the Relying Party prefers an attestation conveyance yielding verifiable
					attestation statements, but allows the client to decide how to obtain such attestation statements.`,
		"none": `Indicates that the Relying Party is not interested in authenticator attestation.`,
	}}

}

type EmailDelivery struct {
	// `enabled` determines whether the API delivers emails.
	// Disable if you want to send the emails yourself. To do so you must subscribe to the `email.create` webhook event.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `from_address` configures the sender address of emails sent to users.
	FromAddress string `yaml:"from_address" json:"from_address,omitempty" koanf:"from_address" split_words:"true" jsonschema:"default=noreply@hanko.io"`
	// `from_name` configures the sender name of emails sent to users.
	FromName string `yaml:"from_name" json:"from_name,omitempty" koanf:"from_name" split_words:"true" jsonschema:"default=Hanko"`
	// `SMTP` contains the SMTP server settings for sending mails.
	SMTP SMTP `yaml:"smtp" json:"smtp,omitempty" koanf:"smtp" jsonschema:"title=smtp"`
}

type Email struct {
	// `acquire_on_login` determines whether users, provided that they do not already have registered an email,
	//	are prompted to provide an email on login.
	AcquireOnLogin bool `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=false"`
	// `acquire_on_registration` determines whether users are prompted to provide an email on registration.
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=true"`
	// `enabled` determines whether emails are enabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// 'limit' determines the maximum number of emails a user can register.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=5"`
	// `max_length` specifies the maximum allowed length of an email address.
	MaxLength int `yaml:"max_length" json:"max_length,omitempty" koanf:"max_length" jsonschema:"default=100"`
	// `optional` determines whether users must provide an email when prompted.
	// There must always be at least one email address associated with an account. The primary email address cannot be
	// deleted if emails are required (`optional`: false`).
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=false"`
	// `passcode_ttl` specifies, in seconds, how long a passcode is valid for.
	PasscodeTtl int `yaml:"passcode_ttl" json:"passcode_ttl,omitempty" koanf:"passcode_ttl" jsonschema:"default=300"`
	// `passlink_ttl` specifies, in seconds, how long a passlink is valid for.
	PasslinkTtl int `yaml:"passlink_ttl" json:"passlink_ttl,omitempty" koanf:"passlink_ttl" jsonschema:"default=300"`
	// `require_verification` determines whether newly created emails must be verified by providing a passcode sent
	// to respective address.
	RequireVerification bool `yaml:"require_verification" json:"require_verification,omitempty" koanf:"require_verification" split_words:"true" jsonschema:"default=true"`
	// `use_as_login_identifier` determines whether emails can be used as an identifier on login.
	UseAsLoginIdentifier bool `yaml:"use_as_login_identifier" json:"use_as_login_identifier,omitempty" koanf:"use_as_login_identifier" jsonschema:"default=true"`
	// `user_for_authentication` determines whether users can log in by providing an email address and subsequently
	// providing a passcode sent to the given email address.
	UseForAuthentication bool `yaml:"use_for_authentication" json:"use_for_authentication,omitempty" koanf:"use_for_authentication" jsonschema:"default=true"`
}

type Username struct {
	// `acquire_on_login` determines whether users, provided that they do not already have set a username,
	//	are prompted to provide a username on login.
	AcquireOnLogin bool `yaml:"acquire_on_login" json:"acquire_on_login,omitempty" koanf:"acquire_on_login" split_words:"true" jsonschema:"default=true"`
	// `acquire_on_registration` determines whether users are prompted to provide a username on registration.
	AcquireOnRegistration bool `yaml:"acquire_on_registration" json:"acquire_on_registration,omitempty" koanf:"acquire_on_registration" split_words:"true" jsonschema:"default=true"`
	// `enabled` determines whether users can set a unique username.
	//
	// Usernames can contain letters (a-z,A-Z), numbers (0-9), and underscores.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `max_length` specifies the maximum allowed length of a username.
	MaxLength int `yaml:"max_length" json:"max_length,omitempty" koanf:"max_length" jsonschema:"default=32"`
	// `min_length` specifies the minimum length of a username.
	MinLength int `yaml:"min_length" json:"min_length,omitempty" koanf:"min_length" split_words:"true" jsonschema:"default=3"`
	// `optional` determines whether users must provide a username when prompted. The username can only be changed but
	// not deleted if usernames are required (`optional: false`).
	Optional bool `yaml:"optional" json:"optional,omitempty" koanf:"optional" jsonschema:"default=true"`
	// `use_as_login_identifier` determines whether usernames, if enabled, can be used for logging in.
	UseAsLoginIdentifier bool `yaml:"use_as_login_identifier" json:"use_as_login_identifier,omitempty" koanf:"use_as_login_identifier" jsonschema:"default=true"`
}

type Passlink struct {
	Enabled bool   `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	URL     string `yaml:"url" json:"url,omitempty" koanf:"url"`
}

func (p *Passlink) Validate() error {
	if len(strings.TrimSpace(p.URL)) == 0 {
		return errors.New("url must not be empty")
	}
	if url, err := url.Parse(p.URL); err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	} else if url.Scheme == "" || url.Host == "" {
		return errors.New("url must be a valid URL")
	}
	return nil
}
