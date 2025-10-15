package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type SamlIDPInitiatedRequestPersister interface {
	Create(samlIDPInitiatedRequest models.SamlIDPInitiatedRequest) error
	GetByResponseIDAndIssuer(responseID, entityID string) (*models.SamlIDPInitiatedRequest, error)
}

type samlIDPInitiatedRequestPersister struct {
	db *pop.Connection
}

func (p samlIDPInitiatedRequestPersister) GetByResponseIDAndIssuer(responseID, entityID string) (*models.SamlIDPInitiatedRequest, error) {
	samlIDPInitiatedRequest := models.SamlIDPInitiatedRequest{}
	query := p.db.Where("response_id = ? AND idp_entity_id = ?", responseID, entityID)
	err := query.First(&samlIDPInitiatedRequest)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	return &samlIDPInitiatedRequest, nil
}

func NewSamlIDPInitiatedRequestPersister(db *pop.Connection) SamlIDPInitiatedRequestPersister {
	return &samlIDPInitiatedRequestPersister{db: db}
}

func (p samlIDPInitiatedRequestPersister) Create(samlIDPInitiatedRequest models.SamlIDPInitiatedRequest) error {
	vErr, err := p.db.ValidateAndCreate(&samlIDPInitiatedRequest)
	if err != nil {
		return fmt.Errorf("failed to store saml idp initiated request: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("saml idp initated request object validation failed: %w", vErr)
	}

	return nil
}
