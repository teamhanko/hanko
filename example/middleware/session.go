package middleware

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"log"
	"net/http"
)

func SessionMiddleware(hankoUrl string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("hanko")
			if err == http.ErrNoCookie {
				return c.Redirect(http.StatusTemporaryRedirect, "/unauthorized")
			}
			if err != nil {
				return err
			}
			set, err := jwk.Fetch(context.Background(), fmt.Sprintf("%v/.well-known/jwks.json", hankoUrl))
			if err != nil {
				return err
			}

			token, err := jwt.Parse([]byte(cookie.Value), jwt.WithKeySet(set))
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/unauthorized")
			}

			log.Printf("session for user '%s' verified successfully", token.Subject())
			c.Set("token", cookie.Value)
			c.Set("user", token.Subject())

			return next(c)
		}
	}
}
