package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SamlStatePersister interface {
	Create(state models.SamlState) error
	GetByNonce(nonce string) (*models.SamlState, error)
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

func (s samlStatePersister) GetByNonce(nonce string) (*models.SamlState, error) {
	state := models.SamlState{}

	err := s.db.Where("nonce = ?", nonce).First(&state)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
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
