package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/config"
	"net/http"
	"testing"
)

func TestThirdParty_GenerateCookie(t *testing.T) {
	cfg := &config.Config{
		Session: config.Session{
			Cookie: config.Cookie{
				Domain:   "lorem",
				HttpOnly: true,
				Secure:   false,
			},
		},
	}

	cookieOptions := CookieOptions{
		MaxAge:   300,
		Path:     "/lorem/cookie",
		SameSite: http.SameSiteLaxMode,
	}

	state := "I am a test state"

	cookie := GenerateStateCookie(cfg, HankoThirdpartyStateCookie, state, cookieOptions)

	assert.NotNil(t, cookie)
	assert.Equal(t, cookie.Name, HankoThirdpartyStateCookie)
	assert.Equal(t, cookie.Value, state)
	assert.Equal(t, cookie.Path, cookieOptions.Path)
	assert.Equal(t, cookie.Domain, cfg.Session.Cookie.Domain)
	assert.Equal(t, cookie.MaxAge, cookieOptions.MaxAge)
	assert.Equal(t, cookie.Secure, cfg.Session.Cookie.Secure)
	assert.Equal(t, cookie.HttpOnly, cfg.Session.Cookie.HttpOnly)
	assert.Equal(t, cookie.SameSite, cookieOptions.SameSite)
}

func TestThirdParty_GenerateCookieWithEmptyPath(t *testing.T) {
	cfg := &config.Config{
		Session: config.Session{
			Cookie: config.Cookie{
				Domain:   "lorem",
				HttpOnly: true,
				Secure:   false,
			},
		},
	}

	cookieOptions := CookieOptions{
		MaxAge:   300,
		SameSite: http.SameSiteLaxMode,
	}

	state := "I am a test state"

	cookie := GenerateStateCookie(cfg, HankoThirdpartyStateCookie, state, cookieOptions)

	assert.Equal(t, cookie.Path, "/")
}

func TestThirdParty_GenerateCookieWithEmptyMaxAge(t *testing.T) {
	cfg := &config.Config{
		Session: config.Session{
			Cookie: config.Cookie{
				Domain:   "lorem",
				HttpOnly: true,
				Secure:   false,
			},
		},
	}

	cookieOptions := CookieOptions{
		MaxAge:   0,
		Path:     "/lorem",
		SameSite: http.SameSiteLaxMode,
	}

	state := "I am a test state"

	cookie := GenerateStateCookie(cfg, HankoThirdpartyStateCookie, state, cookieOptions)

	assert.Equal(t, cookie.MaxAge, 300)
}
