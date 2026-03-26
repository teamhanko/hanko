package handler

import (
	"fmt"
	"net/http"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/session"
)

type SessionHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
	cfg            config.Config
}

func NewSessionHandler(persister persistence.Persister, sessionManager session.Manager, cfg config.Config) *SessionHandler {
	return &SessionHandler{
		persister:      persister,
		sessionManager: sessionManager,
		cfg:            cfg,
	}
}

func (h *SessionHandler) ValidateSession(c echo.Context) error {
	lookup := fmt.Sprintf("header:Authorization:Bearer,cookie:%s", h.cfg.Session.Cookie.GetName())
	extractors, err := echojwt.CreateExtractors(lookup)

	if err != nil {
		return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
	}

	for _, extractor := range extractors {
		auths, extractorErr := extractor(c)
		if extractorErr != nil {
			continue
		}
		for _, auth := range auths {
			token, tokenErr := h.sessionManager.Verify(auth)
			if tokenErr != nil {
				continue
			}

			claims, err := dto.GetClaimsFromToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse token claims: %w", err))
			}

			sessionModel, err := h.persister.GetSessionPersister().Get(claims.SessionID)
			if err != nil {
				return fmt.Errorf("failed to get session from database: %w", err)
			}
			if sessionModel == nil {
				continue
			}

			// Check idle timeout
			idleTimeout, _ := time.ParseDuration(h.cfg.Session.IdleTimeout)
			if idleTimeout > 0 && time.Since(sessionModel.LastUsed) > idleTimeout {
				sessionDeletionErr := h.persister.GetSessionPersister().Delete(*sessionModel)
				if sessionDeletionErr != nil {
					return fmt.Errorf("failed to delete session: %w", sessionDeletionErr)
				}

				cookie, cookieDeletionErr := h.sessionManager.DeleteCookie()
				if cookieDeletionErr != nil {
					return fmt.Errorf("could not delete cookie: %w", cookieDeletionErr)
				}
				c.SetCookie(cookie)

				// session expired due to idle timeout
				continue
			}

			// Update lastUsed field
			sessionModel.LastUsed = time.Now().UTC()
			err = h.persister.GetSessionPersister().Update(*sessionModel)
			if err != nil {
				return dto.ToHttpError(err)
			}

			return c.JSON(http.StatusOK, dto.ValidateSessionResponse{
				IsValid:        true,
				Claims:         claims,
				ExpirationTime: &claims.Expiration,
				UserID:         &claims.Subject,
			})
		}
	}

	return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
}

func (h *SessionHandler) ValidateSessionFromBody(c echo.Context) error {
	var request dto.ValidateSessionRequest
	err := (&echo.DefaultBinder{}).BindBody(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	err = c.Validate(request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	token, err := h.sessionManager.Verify(request.SessionToken)
	if err != nil {
		return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
	}

	claims, err := dto.GetClaimsFromToken(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse token claims: %w", err))
	}

	sessionModel, err := h.persister.GetSessionPersister().Get(claims.SessionID)
	if err != nil {
		return dto.ToHttpError(err)
	}

	if sessionModel == nil {
		return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
	}

	// Check idle timeout
	idleTimeout, _ := time.ParseDuration(h.cfg.Session.IdleTimeout)
	if idleTimeout > 0 && time.Since(sessionModel.LastUsed) > idleTimeout {
		sessionDeletionErr := h.persister.GetSessionPersister().Delete(*sessionModel)
		if sessionDeletionErr != nil {
			return dto.ToHttpError(fmt.Errorf("failed to delete session: %w", sessionDeletionErr))
		}

		cookie, cookieDeletionErr := h.sessionManager.DeleteCookie()
		if cookieDeletionErr != nil {
			return dto.ToHttpError(fmt.Errorf("could not delete cookie: %w", cookieDeletionErr))
		}
		c.SetCookie(cookie)

		// session expired due to idle timeout
		return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
	}

	// update lastUsed field
	sessionModel.LastUsed = time.Now().UTC()
	err = h.persister.GetSessionPersister().Update(*sessionModel)
	if err != nil {
		return dto.ToHttpError(err)
	}

	return c.JSON(http.StatusOK, dto.ValidateSessionResponse{
		IsValid:        true,
		Claims:         claims,
		ExpirationTime: &claims.Expiration,
		UserID:         &claims.Subject,
	})
}
