package config

import (
	"errors"
	"time"
)

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
	// `lifespan` determines the maximum duration for which a session token (JWT) is valid. It must be a (possibly signed) sequence of decimal
	// numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Lifespan string `yaml:"lifespan" json:"lifespan,omitempty" koanf:"lifespan" jsonschema:"default=12h"`
	// `server_side` contains configuration for server-side sessions.
	ServerSide ServerSide `yaml:"server_side" json:"server_side" koanf:"server_side"`
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

type ServerSide struct {
	// `enabled` determines whether server-side sessions are enabled.
	//
	// NOTE: When enabled the session endpoint must be used in order to check if a session is still valid.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `limit` determines the maximum number of server-side sessions a user can have. When the limit is exceeded,
	// older sessions are invalidated.
	Limit int `yaml:"limit" json:"limit,omitempty" koanf:"limit" jsonschema:"default=100"`
}
