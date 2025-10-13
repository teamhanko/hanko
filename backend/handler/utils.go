package handler

import (
	"fmt"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"net/http"
)

func loadDto[I any](ctx echo.Context) (*I, error) {
	var adminDto I
	err := ctx.Bind(&adminDto)
	if err != nil {
		ctx.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = ctx.Validate(adminDto)
	if err != nil {
		ctx.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return &adminDto, nil
}

func storeSession(cfg *config.Config, persister persistence.Persister, userId uuid.UUID, rawToken jwt.Token, httpContext echo.Context, tx *pop.Connection) error {
	activeSessions, err := persister.GetSessionPersisterWithConnection(tx).ListActive(userId)
	if err != nil {
		return fmt.Errorf("failed to list active sessions: %w", err)
	}

	// remove all server side sessions that exceed the limit
	if len(activeSessions) >= cfg.Session.Limit {
		for i := cfg.Session.Limit - 1; i < len(activeSessions); i++ {
			err = persister.GetSessionPersisterWithConnection(tx).Delete(activeSessions[i])
			if err != nil {
				return fmt.Errorf("failed to remove latest session: %w", err)
			}
		}
	}

	sessionID, _ := rawToken.Get("session_id")

	expirationTime := rawToken.Expiration()
	sessionModel := models.Session{
		ID:        uuid.FromStringOrNil(sessionID.(string)),
		UserID:    userId,
		CreatedAt: rawToken.IssuedAt(),
		UpdatedAt: rawToken.IssuedAt(),
		ExpiresAt: &expirationTime,
		LastUsed:  rawToken.IssuedAt(),
	}

	if cfg.Session.AcquireIPAddress {
		sessionModel.IpAddress = nulls.NewString(httpContext.RealIP())
	}

	if cfg.Session.AcquireUserAgent {
		sessionModel.UserAgent = nulls.NewString(httpContext.Request().UserAgent())
	}

	err = persister.GetSessionPersisterWithConnection(tx).Create(sessionModel)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	return nil
}
