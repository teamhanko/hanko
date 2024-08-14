package config

import (
	"errors"
	"fmt"
	"strings"
)

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
