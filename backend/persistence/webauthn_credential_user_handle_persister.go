package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type WebauthnCredentialUserHandlePersister interface {
	GetByHandle(handle string, tenantID uuid.UUID) (*models.WebauthnCredentialUserHandle, error)
}

func NewWebauthnCredentialUserHandlePersister(db *pop.Connection) WebauthnCredentialUserHandlePersister {
	return &webauthnCredentialUserHandlePersister{db: db}
}

type webauthnCredentialUserHandlePersister struct {
	db *pop.Connection
}

func (p *webauthnCredentialUserHandlePersister) GetByHandle(handle string, tenantID uuid.UUID) (*models.WebauthnCredentialUserHandle, error) {
	handleModel := models.WebauthnCredentialUserHandle{}
	query := p.db.Where("handle = ?", handle)
	query = query.Where("tenant_id = ?", tenantID)
	err := query.First(&handleModel)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get handleModel: %w", err)
	}

	return &handleModel, nil
}
