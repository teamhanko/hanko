package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type WebauthnSessionDataPersister interface {
	Get(id uuid.UUID, tenantID *uuid.UUID) (*models.WebauthnSessionData, error)
	GetByChallenge(challenge string, tenantID *uuid.UUID) (*models.WebauthnSessionData, error)
	Create(sessionData models.WebauthnSessionData) error
	Update(sessionData models.WebauthnSessionData) error
	Delete(sessionData models.WebauthnSessionData) error
	Cleanup[models.WebauthnSessionData]
}

type webauthnSessionDataPersister struct {
	db *pop.Connection
}

func NewWebauthnSessionDataPersister(db *pop.Connection) WebauthnSessionDataPersister {
	return &webauthnSessionDataPersister{db: db}
}

func (p *webauthnSessionDataPersister) Get(id uuid.UUID, tenantID *uuid.UUID) (*models.WebauthnSessionData, error) {
	sessionData := models.WebauthnSessionData{}
	query := p.db.Eager().Q()
	if tenantID != nil {
		query = query.Where("webauthn_session_data.tenant_id = ?", tenantID)
	} else {
		query = query.Where("webauthn_session_data.tenant_id IS NULL")
	}
	err := query.Find(&sessionData, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sessionData: %w", err)
	}

	return &sessionData, nil
}

func (p *webauthnSessionDataPersister) GetByChallenge(challenge string, tenantID *uuid.UUID) (*models.WebauthnSessionData, error) {
	var sessionData []models.WebauthnSessionData
	query := p.db.Eager().Where("challenge = ?", challenge)
	if tenantID != nil {
		query = query.Where("webauthn_session_data.tenant_id = ?", tenantID)
	} else {
		query = query.Where("webauthn_session_data.tenant_id IS NULL")
	}
	err := query.All(&sessionData)
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

func (p *webauthnSessionDataPersister) FindExpired(cutoffTime time.Time, page, perPage int, tenantID *uuid.UUID) ([]models.WebauthnSessionData, error) {
	var items []models.WebauthnSessionData

	query := p.db.Where("expires_at < ?", cutoffTime)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	query = query.Select("id").Paginate(page, perPage)
	err := query.All(&items)

	return items, err
}

func (p *webauthnSessionDataPersister) Delete(sessionData models.WebauthnSessionData) error {
	err := p.db.Destroy(&sessionData)
	if err != nil {
		return fmt.Errorf("failed to delete sessionData: %w", err)
	}

	return nil
}
