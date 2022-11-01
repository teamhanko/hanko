package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	"net/http"
	"time"
)

type PrivilegedSessionHandler struct {
	duration time.Duration
}

func NewPrivilegedSession(config config.Session) (*PrivilegedSessionHandler, error) {
	duration, err := time.ParseDuration(config.PrivilegedLifespan)
	if err != nil {
		return nil, fmt.Errorf("failed to parse privileged lifespan: %w", err)
	}
	return &PrivilegedSessionHandler{duration: duration}, nil
}

// Middleware returns a handler that checks whether the session is privileged.
func (h *PrivilegedSessionHandler) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if h.duration.Nanoseconds() == 0 {
			return next(c)
		}

		sessionToken, ok := c.Get("session").(jwt.Token)
		if !ok {
			return c.JSON(http.StatusUnauthorized, nil)
		}

		issuedAt := sessionToken.IssuedAt()
		privilegedUntil := issuedAt.Add(h.duration)

		if time.Now().After(privilegedUntil) {
			return c.JSON(http.StatusUnauthorized, nil)
		}

		return next(c)
	}
}
