package middleware

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
	"time"
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

		if cfg.ServerSide.Enabled {
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

			// Update lastUsed field
			sessionModel.LastUsed = time.Now().UTC()
			err = persister.GetSessionPersister().Update(*sessionModel)
			if err != nil {
				return nil, err
			}
		}

		return token, nil
	}
}
