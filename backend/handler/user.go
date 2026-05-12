package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/session"
	"github.com/teamhanko/hanko/backend/v2/utils"
)

type UserHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
	auditLogger    auditlog.Logger
	cfg            *config.Config
}

func NewUserHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) *UserHandler {
	return &UserHandler{
		persister:      persister,
		auditLogger:    auditLogger,
		sessionManager: sessionManager,
		cfg:            cfg,
	}
}

func (h *UserHandler) Me(c echo.Context) error {
	tenantID, err := utils.TenantIDFromContext(c)
	if err != nil {
		return fmt.Errorf("invalid tenant identifier: %w", err)
	}

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(sessionToken.Subject()), tenantID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	data := dto.ProfileDataFromUserModel(user, h.cfg)
	return c.JSON(http.StatusOK, *data)
}

func (h *UserHandler) Logout(c echo.Context) error {
	tenantID, err := utils.TenantIDFromContext(c)
	if err != nil {
		return fmt.Errorf("invalid tenant identifier: %w", err)
	}

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	userId := uuid.FromStringOrNil(sessionToken.Subject())

	user, err := h.persister.GetUserPersister().Get(userId, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	sID, ok := sessionToken.Get("session_id")
	if ok {
		sessionIDString := sID.(string)
		sessionID, err := uuid.FromString(sessionIDString)
		if err != nil {
			return fmt.Errorf("failed to convert session id to uuid: %w", err)
		}
		sessionModel, err := h.persister.GetSessionPersister().Get(sessionID, tenantID)
		if err != nil {
			return fmt.Errorf("failed to get session from database: %w", err)
		}
		if sessionModel != nil {
			err = h.persister.GetSessionPersister().Delete(*sessionModel)
			if err != nil {
				return fmt.Errorf("failed to delete session from database: %w", err)
			}
		}
	}

	err = h.auditLogger.Create(c, models.AuditLogUserLoggedOut, user, nil, tenantID)
	if err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	cookie, err := h.sessionManager.DeleteCookie()
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	c.SetCookie(cookie)

	return c.NoContent(http.StatusNoContent)
}
