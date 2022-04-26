package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/persistence/models"
)

type PasswordCredentialPersister interface {
	Create(password models.PasswordCredential) error
	GetByUserID(userId uuid.UUID) (*models.PasswordCredential, error)
	Update(password models.PasswordCredential) error
}

type passwordCredentialPersister struct {
	db *pop.Connection
}

func NewPasswordCredentialPersister(db *pop.Connection) PasswordCredentialPersister {
	return &passwordCredentialPersister{db: db}
}

func (p *passwordCredentialPersister) Create(password models.PasswordCredential) error {
	vErr, err := p.db.ValidateAndCreate(&password)
	if err != nil {
		return fmt.Errorf("failed to store password credential: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("password object validation failed: %w", vErr)
	}

	return nil
}

func (p *passwordCredentialPersister) GetByUserID(userId uuid.UUID) (*models.PasswordCredential, error) {
	pw := models.PasswordCredential{}
	query := p.db.Where("user_id = (?)", userId.String())
	err := query.First(&pw)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	return &pw, nil
}

func (p *passwordCredentialPersister) Update(password models.PasswordCredential) error {
	vErr, err := p.db.ValidateAndUpdate(&password)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("password object validation failed: %w", vErr)
	}

	return nil
}
