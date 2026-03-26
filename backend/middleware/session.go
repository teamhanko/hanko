package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/session"
)

// Session is a convenience function to create a middleware.JWT with custom JWT verification
func Session(cfg *config.Config, persister persistence.Persister, generator session.Manager) echo.MiddlewareFunc {
	c := echojwt.Config{
		ContextKey:     "session",
		TokenLookup:    fmt.Sprintf("header:Authorization:Bearer,cookie:%s", cfg.Session.Cookie.GetName()),
		ParseTokenFunc: parseToken(cfg.Session, persister, generator),
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized).SetInternal(err)
		},
	}
	return echojwt.WithConfig(c)
}

type ParseTokenFunc = func(c echo.Context, auth string) (interface{}, error)

func parseToken(cfg config.Session, persister persistence.Persister, generator session.Manager) ParseTokenFunc {
	return func(c echo.Context, auth string) (interface{}, error) {
		token, err := generator.Verify(auth)
		if err != nil {
			return nil, err
		}

		// check that the session id is stored in the database
		sessionId, ok := token.Get("session_id")
		if !ok {
			return nil, errors.New("no session id found in token")
		}
		sessionID, err := uuid.FromString(sessionId.(string))
		if err != nil {
			return nil, errors.New("session id has wrong format")
		}

		sessionModel, err := persister.GetSessionPersister().Get(sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get session from database: %w", err)
		}
		if sessionModel == nil {
			return nil, fmt.Errorf("session id not found in database")
		}

		// Check idle timeout
		idleTimeout, _ := time.ParseDuration(cfg.IdleTimeout)
		if idleTimeout > 0 && time.Since(sessionModel.LastUsed) > idleTimeout {
			sessionDeletionErr := persister.GetSessionPersister().Delete(*sessionModel)
			if sessionDeletionErr != nil {
				return nil, fmt.Errorf("failed to delete session: %w", sessionDeletionErr)
			}

			cookie, cookieDeletionErr := generator.DeleteCookie()
			if cookieDeletionErr != nil {
				return nil, fmt.Errorf("could not delete cookie: %w", cookieDeletionErr)
			}
			c.SetCookie(cookie)

			return nil, errors.New("session expired due to idle timeout")
		}

		// Update lastUsed field
		sessionModel.LastUsed = time.Now().UTC()
		err = persister.GetSessionPersister().Update(*sessionModel)
		if err != nil {
			return nil, err
		}

		return token, nil
	}
}
