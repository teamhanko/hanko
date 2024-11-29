package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

type SessionAdminHandler struct {
	cfg           *config.Config
	persister     persistence.Persister
	sessionManger session.Manager
	auditLogger   auditlog.Logger
}

func NewSessionAdminHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) SessionAdminHandler {
	return SessionAdminHandler{
		cfg:           cfg,
		persister:     persister,
		sessionManger: sessionManager,
		auditLogger:   auditLogger,
	}
}

func (h *SessionAdminHandler) Generate(ctx echo.Context) error {
	var body admin.CreateSessionTokenDto
	if err := (&echo.DefaultBinder{}).BindBody(ctx, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := ctx.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	userID, err := uuid.FromString(body.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	user, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	var emailDTO *dto.EmailJwt
	if email := user.Emails.GetPrimary(); email != nil {
		emailDTO = dto.JwtFromEmailModel(email)
	}

	encodedToken, rawToken, err := h.sessionManger.GenerateJWT(userID, emailDTO)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	if h.cfg.Session.ServerSide.Enabled {
		activeSessions, err := h.persister.GetSessionPersister().ListActive(userID)
		if err != nil {
			return fmt.Errorf("failed to list active sessions: %w", err)
		}

		// remove all server side sessions that exceed the limit
		if len(activeSessions) >= h.cfg.Session.ServerSide.Limit {
			for i := h.cfg.Session.ServerSide.Limit - 1; i < len(activeSessions); i++ {
				err = h.persister.GetSessionPersister().Delete(activeSessions[i])
				if err != nil {
					return fmt.Errorf("failed to remove latest session: %w", err)
				}
			}
		}

		sessionID, _ := rawToken.Get("session_id")

		expirationTime := rawToken.Expiration()
		sessionModel := models.Session{
			ID:        uuid.FromStringOrNil(sessionID.(string)),
			UserID:    userID,
			UserAgent: body.UserAgent,
			IpAddress: body.IpAddress,
			CreatedAt: rawToken.IssuedAt(),
			UpdatedAt: rawToken.IssuedAt(),
			ExpiresAt: &expirationTime,
			LastUsed:  rawToken.IssuedAt(),
		}

		err = h.persister.GetSessionPersister().Create(sessionModel)
		if err != nil {
			return fmt.Errorf("failed to store session: %w", err)
		}
	}

	response := admin.CreateSessionTokenResponse{
		SessionToken: encodedToken,
	}

	err = h.auditLogger.Create(ctx, models.AuditLogLoginSuccess, user, nil, auditlog.Detail("api", "admin"))
	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *SessionAdminHandler) List(ctx echo.Context) error {
	listDto, err := loadDto[admin.ListSessionsRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(listDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	sessions, err := h.persister.GetSessionPersister().ListActive(userID)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, sessions)
}

func (h *SessionAdminHandler) Delete(ctx echo.Context) error {
	deleteDto, err := loadDto[admin.DeleteSessionRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(deleteDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	sessionID, err := uuid.FromString(deleteDto.SessionID)
	if err != nil {
		return fmt.Errorf("failed to parse session_id as uuid: %s", err)
	}

	sessionModel, err := h.persister.GetSessionPersister().Get(sessionID)
	if err != nil {
		return err
	}

	if sessionModel == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	} else if sessionModel.UserID != userID {
		return echo.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("session does not belong to user"))
	}

	err = h.persister.GetSessionPersister().Delete(*sessionModel)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}
