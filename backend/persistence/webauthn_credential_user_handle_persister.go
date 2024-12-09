package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialUserHandlePersister interface {
	GetByHandle(string) (*models.WebauthnCredentialUserHandle, error)
}

func NewWebauthnCredentialUserHandlePersister(db *pop.Connection) WebauthnCredentialUserHandlePersister {
	return &webauthnCredentialUserHandlePersister{db: db}
}

type webauthnCredentialUserHandlePersister struct {
	db *pop.Connection
}

func (p *webauthnCredentialUserHandlePersister) GetByHandle(handle string) (*models.WebauthnCredentialUserHandle, error) {
	handleModel := models.WebauthnCredentialUserHandle{}
	err := p.db.Where("handle = ?", handle).First(&handleModel)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get handleModel: %w", err)
	}

	return &handleModel, nil
}
