package config

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/invopop/jsonschema"
)

type Server struct {
	// `public` contains the server configuration for the public API.
	Public ServerSettings `yaml:"public" json:"public,omitempty" koanf:"public" jsonschema:"title=public"`
	// `admin` contains the server configuration for the admin API.
	Admin ServerSettings `yaml:"admin" json:"admin,omitempty" koanf:"admin" jsonschema:"title=admin"`
	// `management` contains the server configuration for the management API.
	Management ServerSettings `yaml:"management" json:"management,omitempty" koanf:"management" jsonschema:"title=management"`
	// `ip` configures how client IP addresses are resolved. This configuration is global and applies to public, admin,
	// and management APIs.
	IP IPConfig `yaml:"ip" json:"ip,omitempty" koanf:"ip"`
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
	err = s.Management.Validate()
	if err != nil {
		return fmt.Errorf("error validating management server settings: %w", err)
	}
	if err := s.IP.Validate(); err != nil {
		return fmt.Errorf("error validating ip settings: %w", err)
	}
	return nil
}

type ServerSettings struct {
	// `address` is the address of the server to listen on in the form of host:port.
	//
	// See [net.Dial](https://pkg.go.dev/net#Dial) for details of the address format.
	Address string `yaml:"address" json:"address,omitempty" koanf:"address"`
}

type Cors struct {
	// `allow_origins` determines the value of the Access-Control-Allow-Origin
	// response header. This header defines a list of [origins](https://developer.mozilla.org/en-US/docs/Glossary/Origin)
	// that may access the resource.
	//
	// The wildcard characters `*` and `?` are supported and are converted to regex fragments `.*` and `.` accordingly.
	AllowOrigins []string `yaml:"allow_origins" json:"allow_origins" koanf:"allow_origins" split_words:"true" jsonschema:"title=allow_origins,default=http://localhost:8888"`

	// `unsafe_wildcard_origin_allowed` allows a wildcard `*` origin to be used with AllowCredentials
	// flag. In that case we consider any origin allowed and send it back to the client in an `Access-Control-Allow-Origin` header.
	//
	// This is INSECURE and potentially leads to [cross-origin](https://portswigger.net/research/exploiting-cors-misconfigurations-for-bitcoins-and-bounties)
	// attacks. See also https://github.com/labstack/echo/issues/2400 for discussion on the subject.
	//
	// Optional. Default value is `false`.
	UnsafeWildcardOriginAllowed bool `yaml:"unsafe_wildcard_origin_allowed" json:"unsafe_wildcard_origin_allowed" koanf:"unsafe_wildcard_origin_allowed" split_words:"true" jsonschema:"title=unsafe_wildcard_origin_allowed,default=false"`
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
	return nil
}

type IPExtractorType string

const (
	IPExtractorDirect        IPExtractorType = "direct"
	IPExtractorXForwardedFor IPExtractorType = "x_forwarded_for"
	IPExtractorXRealIP       IPExtractorType = "x_real_ip"
)

type IPConfig struct {
	// `extractor` determines how the client IP address is resolved.
	//
	// The default `direct` uses the direct network peer address and ignores forwarding headers.
	// Use `x_forwarded_for` or `x_real_ip` only when Hanko is deployed behind trusted reverse proxies.
	//
	// When using x_forwarded_for or x_real_ip, ensure the application is not directly reachable from the internet and
	// that your reverse proxy strips or overwrites incoming forwarding headers.
	Extractor IPExtractorType `yaml:"extractor" json:"extractor,omitempty" koanf:"extractor" jsonschema:"default=direct,enum=direct,enum=x_forwarded_for,enum=x_real_ip"`

	// `trusted_proxies` contains CIDR ranges of reverse proxies that are trusted to provide client IP headers.
	//
	// Required when `extractor` is `x_forwarded_for` or `x_real_ip`.
	// Do not configure public client IP ranges here. Only configure your own reverse proxies,
	// load balancers, or ingress controllers.
	TrustedProxies []string `yaml:"trusted_proxies" json:"trusted_proxies,omitempty" koanf:"trusted_proxies"`
}

func (IPConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	extractor, _ := schema.Properties.Get("extractor")
	extractor.Extras = map[string]any{"meta:enum": map[string]string{
		"direct": "Uses the direct network peer address. This is the default and ignores X-Forwarded-For / X-Real-IP.",
		"x_forwarded_for": `Uses the X-Forwarded-For header, but only from trusted proxies. Requires trusted_proxies.
							When using this option ensure the application is not directly reachable from the internet
							and that your reverse proxy strips or overwrites incoming forwarding headers.`,
		"x_real_ip": `Uses the X-Real-IP header, but only from trusted proxies. Requires trusted_proxies.
							When using this option ensure the application is not directly reachable from the internet
							and that your reverse proxy strips or overwrites incoming forwarding headers.`,
	}}

	if schema.Extras == nil {
		schema.Extras = map[string]any{}
	}

	// Add if/then manually.
	schema.Extras["if"] = map[string]any{
		"required": []string{"extractor"},
		"properties": map[string]any{
			"extractor": map[string]any{
				"enum": []string{"x_forwarded_for", "x_real_ip"},
			},
		},
	}

	schema.Extras["then"] = map[string]any{
		"required": []string{"trusted_proxies"},
		"properties": map[string]any{
			"trusted_proxies": map[string]any{
				"minItems": 1,
			},
		},
	}
}

func (c IPConfig) Validate() error {
	switch c.Extractor {
	case "", IPExtractorDirect:
		return nil

	case IPExtractorXForwardedFor, IPExtractorXRealIP:
		if len(c.TrustedProxies) == 0 {
			return fmt.Errorf("trusted_proxies must be configured when ip.extractor is %q", c.Extractor)
		}

		for _, trustedProxy := range c.TrustedProxies {
			if _, _, err := net.ParseCIDR(trustedProxy); err != nil {
				return fmt.Errorf("invalid trusted proxy CIDR %q: %w", trustedProxy, err)
			}
		}

		return nil

	default:
		return errors.New("ip.extractor must be one of: direct, x_forwarded_for, x_real_ip")
	}
}
