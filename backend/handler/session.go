package handler

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
	"time"
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
		// TODO: different error?
		return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
	}

	var token jwt.Token
	var lastExtractorErr, lastTokenErr error
	for _, extractor := range extractors {
		auths, extractorErr := extractor(c)
		if extractorErr != nil {
			lastExtractorErr = extractorErr
			continue
		}
		for _, auth := range auths {
			token, tokenErr := h.sessionManager.Verify(auth)
			if tokenErr != nil {
				lastTokenErr = tokenErr
				continue
			}

			if h.cfg.Session.ServerSide.Enabled {
				// check that the session id is stored in the database
				sessionId, ok := token.Get("session_id")
				if !ok {
					lastTokenErr = errors.New("no session id found in token")
					continue
				}
				sessionID, err := uuid.FromString(sessionId.(string))
				if err != nil {
					lastTokenErr = errors.New("session id has wrong format")
					continue
				}

				sessionModel, err := h.persister.GetSessionPersister(nil).Get(sessionID)
				if err != nil {
					return fmt.Errorf("failed to get session from database: %w", err)
				}
				if sessionModel == nil {
					lastTokenErr = fmt.Errorf("session id not found in database")
					continue
				}

				// Update lastUsed field
				sessionModel.LastUsed = time.Now().UTC()
				err = h.persister.GetSessionPersister(nil).Update(*sessionModel)
				if err != nil {
					return dto.ToHttpError(err)
				}
			}

			return nil
		}
	}

	if lastTokenErr != nil {
		return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
	} else if lastExtractorErr != nil {
		return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
	}

	return c.JSON(http.StatusOK, ValidateSessionResponse{
		IsValid:        false,
		ExpirationTime: token.Expiration(),
	})
}
