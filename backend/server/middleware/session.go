package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/session"
)

// Session is a convenience function to create a middleware.JWT with custom JWT verification
func Session(generator session.Manager) echo.MiddlewareFunc {
	c := middleware.JWTConfig{
		ContextKey:     "session",
		TokenLookup:    "header:Authorization,cookie:hanko",
		AuthScheme:     "Bearer",
		ParseTokenFunc: parseToken(generator),
	}
	return middleware.JWTWithConfig(c)
}

type ParseTokenFunc = func(auth string, c echo.Context) (interface{}, error)

func parseToken(generator session.Manager) ParseTokenFunc {
	return func(auth string, c echo.Context) (interface{}, error) {
		return generator.Verify(auth)
	}
}
