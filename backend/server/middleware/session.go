package middleware

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

// Session is a convenience function to create a middleware.JWT with custom JWT verification
func Session(generator session.Manager) echo.MiddlewareFunc {
	c := echojwt.Config{
		ContextKey:     "session",
		TokenLookup:    "header:Authorization:Bearer,cookie:hanko",
		ParseTokenFunc: parseToken(generator),
		ErrorHandler: func(c echo.Context, err error) error {
			return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(err)
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
