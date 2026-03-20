package persistence

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type SamlStatePersister interface {
	Create(state models.SamlState) error
	GetByNonce(nonce string, tenantID *uuid.UUID) (*models.SamlState, error)
	Delete(state models.SamlState) error
}

type samlStatePersister struct {
	db *pop.Connection
}

func NewSamlStatePersister(db *pop.Connection) SamlStatePersister {
	return &samlStatePersister{db: db}
}

func (s samlStatePersister) Create(state models.SamlState) error {
	validationError, err := s.db.ValidateAndCreate(&state)
	if err != nil {
		return err
	}

	if validationError != nil && validationError.HasAny() {
		return fmt.Errorf("token object validation failed: %w", validationError)
	}

	return nil
}

func (s samlStatePersister) GetByNonce(nonce string, tenantID *uuid.UUID) (*models.SamlState, error) {
	state := models.SamlState{}

	query := s.db.Where("nonce = ?", nonce)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.First(&state)
	if err != nil {
		return nil, fmt.Errorf("failed to get state by nonce: %w", err)
	}

	return &state, nil
}

func (s samlStatePersister) Delete(state models.SamlState) error {
	err := s.db.Destroy(&state)
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}

	return nil
}
