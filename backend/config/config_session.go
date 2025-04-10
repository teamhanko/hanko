package config

import (
	"errors"
	"github.com/invopop/jsonschema"
	"time"
)

type Session struct {
	// `allow_revocation` allows users to revoke their own sessions.
	AllowRevocation bool `yaml:"allow_revocation" json:"allow_revocation,omitempty" koanf:"allow_revocation" jsonschema:"default=true"`
	// `audience` is a list of strings that identifies the recipients that the JWT is intended for.
	// The audiences are placed in the `aud` claim of the JWT.
	// If not set, it defaults to the value of the`webauthn.relying_party.id` configuration parameter.
	Audience []string `yaml:"audience" json:"audience,omitempty" koanf:"audience"`
	// `acquire_ip_address` stores the user's IP address in the database.
	AcquireIPAddress bool `yaml:"acquire_ip_address" json:"acquire_ip_address,omitempty" koanf:"acquire_ip_address" jsonschema:"default=true"`
	// `acquire_user_agent` stores the user's user agent in the database.
	AcquireUserAgent bool `yaml:"acquire_user_agent" json:"acquire_user_agent,omitempty" koanf:"acquire_user_agent" jsonschema:"default=true"`
	// `cookie` contains configuration for the session cookie issued on successful registration or login.
	Cookie Cookie `yaml:"cookie" json:"cookie,omitempty" koanf:"cookie"`
	// `enable_auth_token_header` determines whether a session token (JWT) is returned in an `X-Auth-Token`
	// header after a successful authentication. This option should be set to `true` if API and client applications
	// run on different domains.
	EnableAuthTokenHeader bool `yaml:"enable_auth_token_header" json:"enable_auth_token_header,omitempty" koanf:"enable_auth_token_header" split_words:"true" jsonschema:"default=false"`
	// `issuer` is a string that identifies the principal (human user, an organization, or a service)
	// that issued the JWT. Its value is set in the `iss` claim of a JWT.
	Issuer string `yaml:"issuer" json:"issuer,omitempty" koanf:"issuer"`
	// `lifespan` determines the maximum duration for which a session token (JWT) is valid. It must be a (possibly signed) sequence of decimal
	// numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Lifespan string `yaml:"lifespan" json:"lifespan,omitempty" koanf:"lifespan" jsonschema:"default=12h"`
	// `limit` determines the maximum number of server-side sessions a user can have. When the limit is exceeded,
	// older sessions are invalidated.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=5"`
	// `show_on_profile` indicates that the sessions should be listed on the profile.
	ShowOnProfile bool `yaml:"show_on_profile" json:"show_on_profile,omitempty" koanf:"show_on_profile" jsonschema:"default=true"`
	// Deprecated. Use settings in parent object.
	//`server_side` contains configuration for server-side sessions.
	ServerSide *ServerSide `yaml:"server_side" json:"server_side,omitempty" koanf:"server_side"`
}

func (s *Session) Validate() error {
	_, err := time.ParseDuration(s.Lifespan)
	if err != nil {
		return errors.New("failed to parse lifespan")
	}
	return nil
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
	// `retention` determines the retention behavior of authentication cookies.
	Retention string `yaml:"retention" json:"retention,omitempty" koanf:"retention" split_words:"true" jsonschema:"default=persistent,enum=session,enum=persistent,enum=prompt"`
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

func (Cookie) JSONSchemaExtend(schema *jsonschema.Schema) {
	retention, _ := schema.Properties.Get("retention")
	retention.Extras = map[string]any{"meta:enum": map[string]string{
		"session":    "Issues a temporary cookie that lasts for the duration of the browser session.",
		"persistent": "Issues a cookie that remains stored on the user's device until it reaches its expiration date.",
		"prompt":     "Allows the user to choose whether to stay signed in. If the user selects 'Stay signed in', a persistent cookie is issued; a session cookie otherwise.",
	}}
}

func (c *Cookie) GetName() string {
	if c.Name != "" {
		return c.Name
	}

	return "hanko"
}

type ServerSide struct {
	// `enabled` determines whether server-side sessions are enabled.
	//
	// NOTE: When enabled the session endpoint must be used in order to check if a session is still valid.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `limit` determines the maximum number of server-side sessions a user can have. When the limit is exceeded,
	// older sessions are invalidated.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=100"`
}
