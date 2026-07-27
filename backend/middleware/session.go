package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/session"
)

// Session is a convenience function to create a middleware.JWT with custom JWT verification
func Session(persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant, err := context.GetTenant(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get tenant from context")
			}

			jwkManager, err := context.GetJwkManager(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get JWK manager from context")
			}

			sessionManager, err := session.NewManager(jwkManager, tenant.Config)

			echoJwtConfig := echojwt.Config{
				ContextKey:     "session",
				TokenLookup:    fmt.Sprintf("header:Authorization:Bearer,cookie:%s", tenant.Config.Session.Cookie.GetName()),
				ParseTokenFunc: parseToken(*tenant, persister, sessionManager),
				ErrorHandler: func(c echo.Context, err error) error {
					return echo.NewHTTPError(http.StatusUnauthorized).SetInternal(err)
				},
			}

			return echojwt.WithConfig(echoJwtConfig)(next)(c)
		}
	}
}

type ParseTokenFunc = func(c echo.Context, auth string) (interface{}, error)

func parseToken(tenant context.Tenant, persister persistence.Persister, generator session.Manager) ParseTokenFunc {
	return func(c echo.Context, auth string) (interface{}, error) {
		token, err := generator.Verify(auth, tenant.ID)
		if err != nil {
			return nil, err
		}

		sessionId, ok := token.Get("session_id")
		if !ok {
			return nil, errors.New("no session id found in token")
		}
		sessionID, err := uuid.FromString(sessionId.(string))
		if err != nil {
			return nil, errors.New("session id has wrong format")
		}

		sessionModel, err := persister.GetSessionPersister().Get(sessionID, tenant.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get session from database: %w", err)
		}
		if sessionModel == nil {
			return nil, fmt.Errorf("session id not found in database")
		}

		// Check idle timeout
		idleTimeout, _ := time.ParseDuration(tenant.Config.Session.IdleTimeout)
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

func SessionManager() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant, err := context.GetTenant(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get tenant from context")
			}

			jwkManager, err := context.GetJwkManager(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get JWK manager from context")
			}

			sessionManager, err := session.NewManager(jwkManager, tenant.Config)
			c.Set("session_manager", sessionManager)
			return next(c)
		}
	}
}
