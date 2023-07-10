package auditlog

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
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
	CreateWithConnection(*pop.Connection, echo.Context, models.AuditLogType, *models.User, error) error
}

type logger struct {
	persister             persistence.Persister
	storageEnabled        bool
	logger                zeroLog.Logger
	consoleLoggingEnabled bool
}

func NewLogger(persister persistence.Persister, cfg config.AuditLog) Logger {
	var loggerOutput *os.File = nil
	switch cfg.ConsoleOutput.OutputStream {
	case config.OutputStreamStdOut:
		loggerOutput = os.Stdout
	case config.OutputStreamStdErr:
		loggerOutput = os.Stderr
	default:
		loggerOutput = os.Stdout
	}

	return &logger{
		persister:             persister,
		storageEnabled:        cfg.Storage.Enabled,
		logger:                zeroLog.New(loggerOutput),
		consoleLoggingEnabled: cfg.ConsoleOutput.Enabled,
	}
}

func (l *logger) Create(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	return l.CreateWithConnection(l.persister.GetConnection(), context, auditLogType, user, logError)
}

func (l *logger) CreateWithConnection(tx *pop.Connection, context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	if l.storageEnabled {
		err := l.store(tx, context, auditLogType, user, logError)
		if err != nil {
			return err
		}
	}

	if l.consoleLoggingEnabled {
		l.logToConsole(context, auditLogType, user, logError)
	}

	return nil
}

func (l *logger) store(tx *pop.Connection, context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) error {
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create id: %w", err)
	}

	al := models.AuditLog{
		ID:                id,
		Type:              auditLogType,
		Error:             nil,
		MetaHttpRequestId: context.Response().Header().Get(echo.HeaderXRequestID),
		MetaUserAgent:     context.Request().UserAgent(),
		MetaSourceIp:      context.RealIP(),
		ActorUserId:       nil,
		ActorEmail:        nil,
	}

	if user != nil {
		al.ActorUserId = &user.ID
		if e := user.Emails.GetPrimary(); e != nil {
			al.ActorEmail = &e.Address
		}
	}
	if logError != nil {
		// check if error is not nil, because else the string (formatted with fmt.Sprintf) would not be empty but look like this: `%!s(<nil>)`
		tmp := fmt.Sprintf("%s", logError)
		al.Error = &tmp
	}

	return l.persister.GetAuditLogPersisterWithConnection(tx).Create(al)
}

func (l *logger) logToConsole(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error) {
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
		loggerEvent.Str("user_id", user.ID.String())
		if e := user.Emails.GetPrimary(); e != nil {
			loggerEvent.Str("user_email", e.Address)
		}
	}

	loggerEvent.Send()
}
