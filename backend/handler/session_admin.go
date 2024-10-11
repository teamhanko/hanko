package handler

import (
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
	"time"
)

type SessionHandlerAdmin struct {
	persister      persistence.Persister
	sessionManager session.Manager
	cfg            config.Config
}

func NewSessionHandlerAdmin(persister persistence.Persister, sessionManager session.Manager, cfg config.Config) *SessionHandlerAdmin {
	return &SessionHandlerAdmin{
		persister:      persister,
		sessionManager: sessionManager,
		cfg:            cfg,
	}
}

type ValidateSessionRequest struct {
	SessionToken string `json:"session_token"`
}

type ValidateSessionResponse struct {
	IsValid        bool       `json:"is_valid"`
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
}

func (h *SessionHandlerAdmin) ValidateSession(c echo.Context) error {
	var request ValidateSessionRequest
	err := (&echo.DefaultBinder{}).BindBody(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	token, err := h.sessionManager.Verify(request.SessionToken)
	if err != nil {
		return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
	}

	if h.cfg.Session.ServerSide.Enabled {
		// check that the session id is stored in the database
		sessionId, ok := token.Get("session_id")
		if !ok {
			return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
		}
		sessionID, err := uuid.FromString(sessionId.(string))
		if err != nil {
			return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
		}

		sessionModel, err := h.persister.GetSessionPersister(nil).Get(sessionID)
		if err != nil {
			return dto.ToHttpError(err)
		}

		if sessionModel == nil {
			return c.JSON(http.StatusOK, ValidateSessionResponse{IsValid: false})
		}

		// update lastUsed field
		sessionModel.LastUsed = time.Now().UTC()
		err = h.persister.GetSessionPersister(nil).Update(*sessionModel)
		if err != nil {
			return dto.ToHttpError(err)
		}
	}

	expirationTime := token.Expiration()
	userID := uuid.FromStringOrNil(token.Subject())
	return c.JSON(http.StatusOK, ValidateSessionResponse{
		IsValid:        true,
		ExpirationTime: &expirationTime,
		UserID:         &userID,
	})
}
