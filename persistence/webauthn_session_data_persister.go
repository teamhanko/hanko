package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/models"
)

type WebauthnSessionDataPersister struct {
	db *pop.Connection
}

func NewWebauthnSessionDataPersister(db *pop.Connection) *WebauthnSessionDataPersister {
	return &WebauthnSessionDataPersister{db: db}
}

func (p *WebauthnSessionDataPersister) Get(id uuid.UUID) (*models.WebauthnSessionData, error) {
	sessionData := models.WebauthnSessionData{}
	err := p.db.Eager().Find(&sessionData, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sessionData: %w", err)
	}

	return &sessionData, nil
}

func (p WebauthnSessionDataPersister) GetByChallenge(challenge string) (*models.WebauthnSessionData, error) {
	var sessionData []models.WebauthnSessionData
	err := p.db.Eager().Where("challenge = ?", challenge).All(&sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessionData: %w", err)
	}

	if len(sessionData) <= 0 {
		return nil, nil
	}

	return &sessionData[0], nil
}

func (p *WebauthnSessionDataPersister) Create(sessionData models.WebauthnSessionData) error {
	vErr, err := p.db.Eager().ValidateAndCreate(&sessionData)
	if err != nil {
		return fmt.Errorf("failed to store sessionData: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("sessionData object validation failed: %w", vErr)
	}

	return nil
}

func (p *WebauthnSessionDataPersister) Update(sessionData models.WebauthnSessionData) error {
	vErr, err := p.db.Eager().ValidateAndUpdate(&sessionData)
	if err != nil {
		return fmt.Errorf("failed to update sessionData: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("sessionData object validation failed: %w", vErr)
	}

	return nil
}

func (p *WebauthnSessionDataPersister) Delete(sessionData models.WebauthnSessionData) error {
	err := p.db.Eager().Destroy(&sessionData)
	if err != nil {
		return fmt.Errorf("failed to delete sessionData: %w", err)
	}

	return nil
}
