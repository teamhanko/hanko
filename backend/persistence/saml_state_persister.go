package persistence

import (
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type SamlStatePersister interface {
	Create(state models.SamlState) error
	GetByNonce(nonce string, tenantID uuid.UUID) (*models.SamlState, error)
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

// GetByNonce locks the state until the surrounding transaction ends, so it must be called inside one and returns an
// error otherwise.
func (s samlStatePersister) GetByNonce(nonce string, tenantID uuid.UUID) (*models.SamlState, error) {
	if s.db.TX == nil {
		return nil, errors.New("GetByNonce must be called inside a transaction")
	}

	state := models.SamlState{}

	err := s.db.RawQuery("SELECT * FROM saml_states WHERE nonce = ? AND tenant_id = ? FOR UPDATE", nonce, tenantID).First(&state)
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
