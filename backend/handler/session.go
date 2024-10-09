package handler

import (
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
		return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
	}

	var token jwt.Token
	for _, extractor := range extractors {
		auths, extractorErr := extractor(c)
		if extractorErr != nil {
			continue
		}
		for _, auth := range auths {
			t, tokenErr := h.sessionManager.Verify(auth)
			if tokenErr != nil {
				continue
			}

			if h.cfg.Session.ServerSide.Enabled {
				// check that the session id is stored in the database
				sessionId, ok := t.Get("session_id")
				if !ok {
					continue
				}
				sessionID, err := uuid.FromString(sessionId.(string))
				if err != nil {
					continue
				}

				sessionModel, err := h.persister.GetSessionPersister(nil).Get(sessionID)
				if err != nil {
					return fmt.Errorf("failed to get session from database: %w", err)
				}
				if sessionModel == nil {
					continue
				}

				// Update lastUsed field
				sessionModel.LastUsed = time.Now().UTC()
				err = h.persister.GetSessionPersister(nil).Update(*sessionModel)
				if err != nil {
					return dto.ToHttpError(err)
				}
			}

			token = t
			break
		}
	}

	if token != nil {
		expirationTime := token.Expiration()
		return c.JSON(http.StatusOK, ValidateSessionResponse{
			IsValid:        true,
			ExpirationTime: &expirationTime,
		})
	} else {
		return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
	}
}
