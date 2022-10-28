package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnSessionDataPersister interface {
	Get(id uuid.UUID) (*models.WebauthnSessionData, error)
	GetByChallenge(challenge string) (*models.WebauthnSessionData, error)
	Create(sessionData models.WebauthnSessionData) error
	Update(sessionData models.WebauthnSessionData) error
	Delete(sessionData models.WebauthnSessionData) error
}

type webauthnSessionDataPersister struct {
	db *pop.Connection
}

func NewWebauthnSessionDataPersister(db *pop.Connection) WebauthnSessionDataPersister {
	return &webauthnSessionDataPersister{db: db}
}

func (p *webauthnSessionDataPersister) Get(id uuid.UUID) (*models.WebauthnSessionData, error) {
	sessionData := models.WebauthnSessionData{}
	err := p.db.Eager().Find(&sessionData, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sessionData: %w", err)
	}

	return &sessionData, nil
}

func (p *webauthnSessionDataPersister) GetByChallenge(challenge string) (*models.WebauthnSessionData, error) {
	var sessionData []models.WebauthnSessionData
	err := p.db.Eager().Where("challenge = ?", challenge).All(&sessionData)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sessionData: %w", err)
	}

	if len(sessionData) <= 0 {
		return nil, nil
	}

	return &sessionData[0], nil
}

func (p *webauthnSessionDataPersister) Create(sessionData models.WebauthnSessionData) error {
	vErr, err := p.db.Eager().ValidateAndCreate(&sessionData)
	if err != nil {
		return fmt.Errorf("failed to store sessionData: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("sessionData object validation failed: %w", vErr)
	}

	return nil
}

func (p *webauthnSessionDataPersister) Update(sessionData models.WebauthnSessionData) error {
	vErr, err := p.db.Eager().ValidateAndUpdate(&sessionData)
	if err != nil {
		return fmt.Errorf("failed to update sessionData: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("sessionData object validation failed: %w", vErr)
	}

	return nil
}

func (p *webauthnSessionDataPersister) Delete(sessionData models.WebauthnSessionData) error {
	err := p.db.Destroy(&sessionData)
	if err != nil {
		return fmt.Errorf("failed to delete sessionData: %w", err)
	}

	return nil
}
