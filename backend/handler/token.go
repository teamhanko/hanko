package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	rateLimit "github.com/teamhanko/hanko/backend/rate_limiter"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
	"time"
)

type TokenHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
	cfg            *config.Config
	auditLogger    auditlog.Logger
	rateLimiter    limiter.Store
}

func NewTokenHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) *TokenHandler {
	var rateLimiter limiter.Store
	if cfg.RateLimiter.Enabled {
		rateLimiter = rateLimit.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.TokenLimits)
	}

	return &TokenHandler{cfg: cfg,
		persister:      persister,
		sessionManager: sessionManager,
		auditLogger:    auditLogger,
		rateLimiter:    rateLimiter,
	}
}

type TokenValidationBody struct {
	Value string `json:"value" validate:"required"`
}

func (h TokenHandler) Validate(c echo.Context) error {
	if h.rateLimiter != nil {
		err := rateLimit.Limit(h.rateLimiter, uuid.Nil, c)
		if err != nil {
			return err
		}
	}

	var userID uuid.UUID
	err := h.persister.Transaction(func(tx *pop.Connection) error {
		var body TokenValidationBody
		if terr := (&echo.DefaultBinder{}).BindBody(c, &body); terr != nil {
			return dto.ToHttpError(terr)
		}

		if terr := c.Validate(body); terr != nil {
			return dto.ToHttpError(terr)
		}

		tokenPersister := h.persister.GetTokenPersisterWithConnection(tx)
		token, terr := tokenPersister.GetByValue(body.Value)
		if terr != nil {
			return fmt.Errorf("failed to fetch token from db: %w", terr)
		}

		if token == nil {
			return echo.NewHTTPError(http.StatusNotFound, "token not found")
		}

		if time.Now().UTC().After(token.ExpiresAt) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "token has expired")
		}

		terr = tokenPersister.Delete(*token)
		if terr != nil {
			return fmt.Errorf("failed to delete token from db: %w", terr)
		}

		err := h.sessionManager.GenerateCookieOrHeader(token.UserID, c)
		if err != nil {
			return fmt.Errorf("failed to generate cookie or header: %w", err)
		}

		userID = token.UserID

		return nil
	})

	if err != nil {
		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			aerr := h.auditLogger.Create(c, models.AuditLogTokenExchangeFailed, nil, err)
			if aerr != nil {
				return fmt.Errorf("could not create audit log: %w", aerr)
			}
		}
		return err
	}

	user := &models.User{ID: userID}
	err = h.auditLogger.Create(c, models.AuditLogTokenExchangeSucceeded, user, nil)
	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]string{"user_id": userID.String()})

}
