package handler

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	auditlog "github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type SessionAdminHandler struct {
	persister   persistence.Persister
	auditLogger auditlog.Logger
}

func NewSessionAdminHandler(persister persistence.Persister, auditLogger auditlog.Logger) SessionAdminHandler {
	return SessionAdminHandler{
		persister:   persister,
		auditLogger: auditLogger,
	}
}

func (h *SessionAdminHandler) Generate(ctx echo.Context) error {
	tenant, err := context.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	sessionManager, err := context.GetSessionManager(ctx)
	if err != nil {
		return fmt.Errorf("failed to get session manager from context: %w", err)
	}

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

	user, err := h.persister.GetUserPersister().Get(userID, tenant.ID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	encodedToken, rawToken, err := sessionManager.GenerateJWT(dto.UserJWTFromUserModel(user), tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	activeSessions, err := h.persister.GetSessionPersister().ListActive(userID, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to list active sessions: %w", err)
	}

	// remove all server side sessions that exceed the limit
	if len(activeSessions) >= tenant.Config.Session.Limit {
		for i := tenant.Config.Session.Limit - 1; i < len(activeSessions); i++ {
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
		TenantID:  tenant.ID,
		UserID:    userID,
		CreatedAt: rawToken.IssuedAt(),
		UpdatedAt: rawToken.IssuedAt(),
		ExpiresAt: &expirationTime,
		LastUsed:  rawToken.IssuedAt(),
	}

	if len(body.UserAgent) > 0 {
		sessionModel.UserAgent = nulls.NewString(body.UserAgent)
	}

	if len(body.IpAddress) > 0 {
		sessionModel.IpAddress = nulls.NewString(body.IpAddress)
	}

	err = h.persister.GetSessionPersister().Create(sessionModel)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	response := admin.CreateSessionTokenResponse{
		SessionToken: encodedToken,
	}

	err = h.auditLogger.Create(ctx, models.AuditLogLoginSuccess, user, nil, tenant.ID, auditlog.Detail("api", "admin"))
	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *SessionAdminHandler) List(ctx echo.Context) error {
	tenant, err := context.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	listDto, err := loadDto[admin.ListSessionsRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(listDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID, tenant.ID)
	if err != nil {
		return err
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	sessions, err := h.persister.GetSessionPersister().ListActive(userID, tenant.ID)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, sessions)
}

func (h *SessionAdminHandler) Delete(ctx echo.Context) error {
	tenant, err := context.GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	deleteDto, err := loadDto[admin.DeleteSessionRequestDto](ctx)
	if err != nil {
		return err
	}

	userID, err := uuid.FromString(deleteDto.UserID)
	if err != nil {
		return fmt.Errorf(parseUserUuidFailureMessage, err)
	}

	user, err := h.persister.GetUserPersister().Get(userID, tenant.ID)
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

	sessionModel, err := h.persister.GetSessionPersister().Get(sessionID, tenant.ID)
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
