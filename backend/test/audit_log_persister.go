package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewAuditLogPersister(init []models.AuditLog) persistence.AuditLogPersister {
	if init == nil {
		return &auditLogPersister{[]models.AuditLog{}}
	}
	return &auditLogPersister{append([]models.AuditLog{}, init...)}
}

type auditLogPersister struct {
	logs []models.AuditLog
}

func (p *auditLogPersister) Create(auditLog models.AuditLog) error {
	p.logs = append(p.logs, auditLog)
	return nil
}

func (p *auditLogPersister) Get(id uuid.UUID) (*models.AuditLog, error) {
	var found *models.AuditLog
	for _, data := range p.logs {
		if data.ID == id {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (p *auditLogPersister) List(page int, perPage int) ([]models.AuditLog, error) {
	if len(p.logs) == 0 {
		return p.logs, nil
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	var result [][]models.AuditLog
	var j int
	for i := 0; i < len(p.logs); i += perPage {
		j += perPage
		if j > len(p.logs) {
			j = len(p.logs)
		}
		result = append(result, p.logs[i:j])
	}

	if page > len(result) {
		return []models.AuditLog{}, nil
	}
	return result[page-1], nil
}

func (p *auditLogPersister) Delete(auditLog models.AuditLog) error {
	index := -1
	for i, log := range p.logs {
		if log.ID == auditLog.ID {
			index = i
		}
	}
	if index > -1 {
		p.logs = append(p.logs[:index], p.logs[index+1:]...)
	}

	return nil
}
