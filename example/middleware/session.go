package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"log"
	"net/http"
)

func SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("hanko")
			if err == http.ErrNoCookie {
				return c.Redirect(http.StatusTemporaryRedirect, "/unauthorized")
			}
			if err != nil {
				return err
			}
			set, err := jwk.Fetch(context.Background(), "http://hanko:8000/.well-known/jwks.json")
			if err != nil {
				return err
			}

			token, err := jwt.Parse([]byte(cookie.Value), jwt.WithKeySet(set))
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/unauthorized")
			}

			log.Printf("session for user '%s' verified successfully", token.Subject())

			return next(c)
		}
	}
}
