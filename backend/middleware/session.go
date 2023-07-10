package middleware

import (
	"fmt"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

// Session is a convenience function to create a middleware.JWT with custom JWT verification
func Session(cfg *config.Config, generator session.Manager) echo.MiddlewareFunc {
	c := echojwt.Config{
		ContextKey:     "session",
		TokenLookup:    fmt.Sprintf("header:Authorization:Bearer,cookie:%s", cfg.Session.Cookie.GetName()),
		ParseTokenFunc: parseToken(generator),
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized).SetInternal(err)
		},
	}
	return echojwt.WithConfig(c)
}

type ParseTokenFunc = func(c echo.Context, auth string) (interface{}, error)

func parseToken(generator session.Manager) ParseTokenFunc {
	return func(c echo.Context, auth string) (interface{}, error) {
		return generator.Verify(auth)
	}
}
