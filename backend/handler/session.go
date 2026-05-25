package handler

import (
	"fmt"
	"net/http"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/context"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

type SessionHandler struct {
	persister persistence.Persister
}

func NewSessionHandler(persister persistence.Persister) *SessionHandler {
	return &SessionHandler{
		persister: persister,
	}
}

func (h *SessionHandler) ValidateSession(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	sessionManager, err := context.GetSessionManager(c)
	if err != nil {
		return fmt.Errorf("failed to get session manager from context: %w", err)
	}

	lookup := fmt.Sprintf("header:Authorization:Bearer,cookie:%s", tenant.Config.Session.Cookie.GetName())
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
			token, tokenErr := sessionManager.Verify(auth, tenant.ID)
			if tokenErr != nil {
				continue
			}

			claims, err := dto.GetClaimsFromToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse token claims: %w", err))
			}

			sessionModel, err := h.persister.GetSessionPersister().Get(claims.SessionID, tenant.ID)
			if err != nil {
				return fmt.Errorf("failed to get session from database: %w", err)
			}
			if sessionModel == nil {
				continue
			}

			// Check idle timeout
			idleTimeout, _ := time.ParseDuration(tenant.Config.Session.IdleTimeout)
			if idleTimeout > 0 && time.Since(sessionModel.LastUsed) > idleTimeout {
				sessionDeletionErr := h.persister.GetSessionPersister().Delete(*sessionModel)
				if sessionDeletionErr != nil {
					return fmt.Errorf("failed to delete session: %w", sessionDeletionErr)
				}

				cookie, cookieDeletionErr := sessionManager.DeleteCookie()
				if cookieDeletionErr != nil {
					return fmt.Errorf("could not delete cookie: %w", cookieDeletionErr)
				}
				c.SetCookie(cookie)

				// session expired due to idle timeout
				continue
			}

			var idleExpiresAt *time.Time
			if idleTimeout > 0 {
				expiresAt := sessionModel.LastUsed.Add(idleTimeout)
				// Don't exceed JWT expiration
				if expiresAt.After(claims.Expiration) {
					expiresAt = claims.Expiration
				}
				idleExpiresAt = &expiresAt
			}

			return c.JSON(http.StatusOK, dto.ValidateSessionResponse{
				IsValid:        true,
				Claims:         claims,
				ExpirationTime: &claims.Expiration,
				UserID:         &claims.Subject,
				IdleExpiresAt:  idleExpiresAt,
			})
		}
	}

	return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
}

func (h *SessionHandler) ValidateSessionFromBody(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	sessionManager, err := context.GetSessionManager(c)
	if err != nil {
		return fmt.Errorf("failed to get session manager from context: %w", err)
	}

	var request dto.ValidateSessionRequest
	err = (&echo.DefaultBinder{}).BindBody(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	err = c.Validate(request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	token, err := sessionManager.Verify(request.SessionToken, tenant.ID)
	if err != nil {
		return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
	}

	claims, err := dto.GetClaimsFromToken(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse token claims: %w", err))
	}

	sessionModel, err := h.persister.GetSessionPersister().Get(claims.SessionID, tenant.ID)
	if err != nil {
		return dto.ToHttpError(err)
	}

	if sessionModel == nil {
		return c.JSON(http.StatusOK, dto.ValidateSessionResponse{IsValid: false})
	}

	// Check idle timeout
	idleTimeout, _ := time.ParseDuration(tenant.Config.Session.IdleTimeout)
	if idleTimeout > 0 && time.Since(sessionModel.LastUsed) > idleTimeout {
		sessionDeletionErr := h.persister.GetSessionPersister().Delete(*sessionModel)
		if sessionDeletionErr != nil {
			return dto.ToHttpError(fmt.Errorf("failed to delete session: %w", sessionDeletionErr))
		}

		cookie, cookieDeletionErr := sessionManager.DeleteCookie()
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

	var idleExpiresAt *time.Time
	if idleTimeout > 0 {
		expiresAt := sessionModel.LastUsed.Add(idleTimeout)
		// Don't exceed JWT expiration
		if expiresAt.After(claims.Expiration) {
			expiresAt = claims.Expiration
		}
		idleExpiresAt = &expiresAt
	}

	return c.JSON(http.StatusOK, dto.ValidateSessionResponse{
		IsValid:        true,
		Claims:         claims,
		ExpirationTime: &claims.Expiration,
		UserID:         &claims.Subject,
		IdleExpiresAt:  idleExpiresAt,
	})
}
