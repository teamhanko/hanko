package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type AuditLogPersister interface {
	Create(auditLog models.AuditLog) error
	Get(id uuid.UUID) (*models.AuditLog, error)
	List(page int, perPage int, startTime *time.Time, endTime *time.Time) ([]models.AuditLog, error)
	Delete(auditLog models.AuditLog) error
	Count(startTime *time.Time, endTime *time.Time) (int, error)
}

type auditLogPersister struct {
	db *pop.Connection
}

func NewAuditLogPersister(db *pop.Connection) AuditLogPersister {
	return &auditLogPersister{db: db}
}

func (p *auditLogPersister) Create(auditLog models.AuditLog) error {
	vErr, err := p.db.ValidateAndCreate(&auditLog)
	if err != nil {
		return fmt.Errorf("failed to store auditlog: %w", err)
	}
	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("auditlog object validation failed: %w", vErr)
	}

	return nil
}

func (p *auditLogPersister) Get(id uuid.UUID) (*models.AuditLog, error) {
	auditLog := models.AuditLog{}
	err := p.db.Eager().Find(&auditLog, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auditlog: %w", err)
	}

	return &auditLog, nil
}

func (p *auditLogPersister) List(page int, perPage int, startTime *time.Time, endTime *time.Time) ([]models.AuditLog, error) {
	auditLogs := []models.AuditLog{}

	query := p.db.Q()
	if startTime != nil {
		query = query.Where("created_at > ?", startTime)
	}
	if endTime != nil {
		query = query.Where("created_at < ?", endTime)
	}
	err := query.Paginate(page, perPage).Order("created_at desc").All(&auditLogs)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return auditLogs, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch auditLogs: %w", err)
	}

	return auditLogs, nil
}

func (p *auditLogPersister) Delete(auditLog models.AuditLog) error {
	err := p.db.Eager().Destroy(&auditLog)
	if err != nil {
		return fmt.Errorf("failed to delete auditlog: %w", err)
	}

	return nil
}

func (p *auditLogPersister) Count(startTime *time.Time, endTime *time.Time) (int, error) {
	query := p.db.Q()
	if startTime != nil {
		query = query.Where("created_at > ?", startTime)
	}
	if endTime != nil {
		query = query.Where("created_at < ?", endTime)
	}
	count, err := query.Count(&models.AuditLog{})
	if err != nil {
		return 0, fmt.Errorf("failed to get auditLog count: %w", err)
	}

	return count, nil
}
