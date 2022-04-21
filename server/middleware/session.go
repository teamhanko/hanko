package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/session"
)

func Session(generator session.Generator) echo.MiddlewareFunc {
	c := middleware.JWTConfig{
		ContextKey:     "session",
		TokenLookup:    "header:Authorization,cookie:hanko",
		AuthScheme:     "Bearer",
		ParseTokenFunc: parseToken(generator),
	}
	return middleware.JWTWithConfig(c)
}

type ParseTokenFunc = func(auth string, c echo.Context) (interface{}, error)

func parseToken(generator session.Generator) ParseTokenFunc {
	return func(auth string, c echo.Context) (interface{}, error) {
		return generator.Verify(auth)
	}
}
