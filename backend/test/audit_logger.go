package test

import (
	"sync"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// AuditLogEntry is an audit log entry that an AuditLogger was asked to create.
type AuditLogEntry struct {
	Type     models.AuditLogType
	User     *models.User
	Error    error
	TenantID uuid.UUID
}

// AuditLogger is an audit logger that stores nothing and records the entries it is asked to create.
type AuditLogger struct {
	mutex   sync.Mutex
	entries []AuditLogEntry
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

// Entries returns the entries recorded so far.
func (a *AuditLogger) Entries() []AuditLogEntry {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return append([]AuditLogEntry(nil), a.entries...)
}

func (a *AuditLogger) Create(context echo.Context, logType models.AuditLogType, user *models.User, err error, tenantID uuid.UUID, opts ...auditlog.DetailOption) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.entries = append(a.entries, AuditLogEntry{Type: logType, User: user, Error: err, TenantID: tenantID})
	return nil
}

func (a *AuditLogger) CreateWithConnection(tx *pop.Connection, context echo.Context, logType models.AuditLogType, user *models.User, err error, tenantID uuid.UUID, opts ...auditlog.DetailOption) error {
	return a.Create(context, logType, user, err, tenantID, opts...)
}
