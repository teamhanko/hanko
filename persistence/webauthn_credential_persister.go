package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/persistence/models"
)

type WebauthnCredentialPersister struct {
	db *pop.Connection
}

func NewWebauthnCredentialPersister(db *pop.Connection) *WebauthnCredentialPersister {
	return &WebauthnCredentialPersister{db: db}
}

func (p *WebauthnCredentialPersister) Get(id string) (*models.WebauthnCredential, error) {
	credential := models.WebauthnCredential{}
	err := p.db.Find(&credential, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	return &credential, nil
}

func (p *WebauthnCredentialPersister) Create(credential models.WebauthnCredential) error {
	vErr, err := p.db.ValidateAndCreate(&credential)
	if err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("credential object validation failed: %w", vErr)
	}

	return nil
}

func (p *WebauthnCredentialPersister) Update(credential models.WebauthnCredential) error {
	vErr, err := p.db.ValidateAndUpdate(&credential)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("credential object validation failed: %w", vErr)
	}

	return nil
}

func (p *WebauthnCredentialPersister) Delete(credential models.WebauthnCredential) error {
	err := p.db.Destroy(&credential)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	return nil
}
