package auditlog

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"strconv"
	"time"
)

type Client interface {
	Create(echo.Context, models.AuditLogType, *models.User, error) error
}

type client struct {
	persister      persistence.Persister
	storageEnabled bool
}

func NewClient(persister persistence.Persister, config config.AuditLog) Client {
	return &client{
		persister:      persister,
		storageEnabled: config.Storage.Enabled,
	}
}

func (c *client) Create(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	var err error = nil
	if c.storageEnabled {
		err = c.store(context, auditLogType, user, logError)
		if err != nil {
			return err
		}
	}

	now := time.Now()
	loggerEvent := zeroLogger.Log().
		Str("audience", "audit").
		Str("type", string(auditLogType)).
		AnErr("error", logError).
		Str("http_request_id", context.Response().Header().Get(echo.HeaderXRequestID)).
		Str("source_ip", context.RealIP()).
		Str("user_agent", context.Request().UserAgent()).
		Str("time", now.Format(time.RFC3339Nano)).
		Str("time_unix", strconv.FormatInt(now.Unix(), 10))

	if user != nil {
		loggerEvent.Str("user_id", user.ID.String()).
			Str("user_email", user.Email)
	}

	loggerEvent.Send()

	return nil
}

func (c *client) store(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create id: %w", err)
	}
	var userId *uuid.UUID = nil
	var userEmail *string = nil
	if user != nil {
		userId = &user.ID
		userEmail = &user.Email
	}
	var errString *string = nil
	if logError != nil {
		// check if error is not nil, because else the string (formatted with fmt.Sprintf) would not be empty but look like this: `%!s(<nil>)`
		tmp := fmt.Sprintf("%s", logError)
		errString = &tmp
	}
	e := models.AuditLog{
		ID:                id,
		Type:              auditLogType,
		Error:             errString,
		MetaHttpRequestId: context.Response().Header().Get(echo.HeaderXRequestID),
		MetaUserAgent:     context.Request().UserAgent(),
		MetaSourceIp:      context.RealIP(),
		ActorUserId:       userId,
		ActorEmail:        userEmail,
	}

	return c.persister.GetAuditLogPersister().Create(e)
}
