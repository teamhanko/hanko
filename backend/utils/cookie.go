package utils

import (
	"github.com/teamhanko/hanko/backend/config"
	"net/http"
)

const (
	HankoThirdpartyStateCookie = "hanko_thirdparty_state"
	HankoTokenQuery            = "hanko_token"
)

type CookieOptions struct {
	MaxAge   int
	Path     string
	SameSite http.SameSite
}

func GenerateStateCookie(config *config.Config, name string, state string, options CookieOptions) *http.Cookie {
	if options.Path == "" {
		options.Path = "/"
	}

	if options.MaxAge == 0 {
		options.MaxAge = 300
	}

	return &http.Cookie{
		Name:     name,
		Value:    state,
		Path:     options.Path,
		Domain:   config.Session.Cookie.Domain,
		MaxAge:   options.MaxAge,
		Secure:   config.Session.Cookie.Secure,
		HttpOnly: config.Session.Cookie.HttpOnly,
		SameSite: options.SameSite,
	}
}
