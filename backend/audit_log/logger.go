package auditlog

import (
	"github.com/gobuffalo/pop/v6"
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
	Create(echo.Context, models.AuditLogType, *models.User, error, ...DetailOption) error
	CreateWithConnection(*pop.Connection, echo.Context, models.AuditLogType, *models.User, error, ...DetailOption) error
}

type logger struct {
	persister             persistence.Persister
	storageEnabled        bool
	logger                zeroLog.Logger
	consoleLoggingEnabled bool
	mask                  bool
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
		mask:                  cfg.Mask,
	}
}

type DetailOption func(map[string]interface{})

func Detail(key string, value interface{}) DetailOption {
	return func(d map[string]interface{}) {
		if value != "" || value != nil {
			d[key] = value
		}
	}
}

func (l *logger) Create(context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error, detailOpts ...DetailOption) error {
	return l.CreateWithConnection(l.persister.GetConnection(), context, auditLogType, user, logError, detailOpts...)
}

func (l *logger) CreateWithConnection(tx *pop.Connection, context echo.Context, auditLogType models.AuditLogType, user *models.User, logError error, detailOpts ...DetailOption) error {
	details := make(map[string]interface{})
	for _, detailOpt := range detailOpts {
		detailOpt(details)
	}

	auditLog, err := models.NewAuditLog(auditLogType, l.getRequestMeta(context), details, user, logError)
	if err != nil {
		return err
	}

	if l.mask {
		auditLog = auditLog.Mask()
	}

	if l.storageEnabled {
		err = l.store(tx, auditLog)
		if err != nil {
			return err
		}
	}

	if l.consoleLoggingEnabled {
		l.logToConsole(auditLog)
	}

	return nil
}

func (l *logger) store(tx *pop.Connection, auditLog models.AuditLog) error {
	return l.persister.GetAuditLogPersisterWithConnection(tx).Create(auditLog)
}

func (l *logger) logToConsole(auditLog models.AuditLog) {
	var err string
	if auditLog.Error != nil {
		err = *auditLog.Error
	}

	now := time.Now()
	loggerEvent := zeroLogger.Log().
		Str("audience", "audit").
		Str("type", string(auditLog.Type)).
		Str("error", err).
		Str("http_request_id", auditLog.MetaHttpRequestId).
		Str("source_ip", auditLog.MetaSourceIp).
		Str("user_agent", auditLog.MetaUserAgent).
		Any("details", auditLog.Details).
		Str("time", now.Format(time.RFC3339Nano)).
		Str("time_unix", strconv.FormatInt(now.Unix(), 10))

	if auditLog.ActorUserId != nil {
		loggerEvent.Str("user_id", auditLog.ActorUserId.String())
		if auditLog.ActorEmail != nil {
			loggerEvent.Str("user_email", *auditLog.ActorEmail)
		}
	}

	loggerEvent.Send()
}

func (l *logger) getRequestMeta(c echo.Context) models.RequestMeta {
	return models.RequestMeta{
		HttpRequestId: c.Response().Header().Get(echo.HeaderXRequestID),
		UserAgent:     c.Request().UserAgent(),
		SourceIp:      c.RealIP(),
	}
}
