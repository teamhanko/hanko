package test

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewAuditLogger() auditlog.Logger {
	return &auditLogger{}
}

type auditLogger struct {
}

func (a *auditLogger) Create(context echo.Context, logType models.AuditLogType, user *models.User, err error) error {
	return nil
}

func (a *auditLogger) CreateWithConnection(tx *pop.Connection, context echo.Context, logType models.AuditLogType, user *models.User, err error) error {
	return nil
}
