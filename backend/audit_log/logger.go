package auditlog

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	zeroLog "github.com/rs/zerolog"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"os"
	"strconv"
	"time"
)

type Logger interface {
	Create(echo.Context, models.AuditLogType, *models.User, error) error
}

type logger struct {
	persister          persistence.Persister
	storageEnabled     bool
	logSensitiveValues bool
	logger             zeroLog.Logger
	enabled            bool
}

func NewLogger(persister persistence.Persister, cfg config.AuditLog) Logger {
	var loggerOutput *os.File = nil
	switch cfg.Output {
	case config.OutputStreamStdOut:
		loggerOutput = os.Stdout
	case config.OutputStreamStdErr:
		loggerOutput = os.Stderr
	default:
		loggerOutput = os.Stdout
	}
	return &logger{
		persister:          persister,
		storageEnabled:     cfg.Storage.Enabled,
		logSensitiveValues: cfg.LogSensitiveValues,
		logger:             zeroLog.New(loggerOutput),
		enabled:            cfg.Enabled,
	}
}

func (c *logger) Create(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	if !c.enabled {
		return nil
	}
	var err error = nil
	if c.storageEnabled {
		err = c.store(context, auditLogType, user, logError)
		if err != nil {
			return err
		}
	}

	now := time.Now()
	loggerEvent := zeroLogger.Log().
		//loggerEvent := c.logger.Log().
		Str("audience", "audit").
		Str("type", string(auditLogType)).
		AnErr("error", logError).
		Str("http_request_id", context.Response().Header().Get(echo.HeaderXRequestID)).
		Str("source_ip", c.maskSensitiveValue(context.RealIP())).
		Str("user_agent", c.maskSensitiveValue(context.Request().UserAgent())).
		Str("time", now.Format(time.RFC3339Nano)).
		Str("time_unix", strconv.FormatInt(now.Unix(), 10))

	if user != nil {
		loggerEvent.Str("user_id", c.maskSensitiveValue(user.ID.String())).
			Str("user_email", c.maskSensitiveValue(user.Email))
	}

	loggerEvent.Send()

	return nil
}

// {"audience":"audit","type":"webauthn_authentication_final_succeeded","http_request_id":"AqV2taczMZk6ERSjhv3XARayUEsFwGzZ","source_ip":"1*****1","user_agent":"M*****6","time":"2022-10-11T14:10:20.5859624Z","time_unix":"1665497420","user_id":"6*****9","user_email":"f*****o"}

func (c *logger) store(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
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

func (c *logger) maskSensitiveValue(value string) string {
	if c.logSensitiveValues {
		return value
	}
	if value == "" {
		return ""
	}

	firstCharacter := value[:1]
	lastCharacter := value[len(value)-1:]

	return fmt.Sprintf("%s*****%s", firstCharacter, lastCharacter)
}
