package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialPersister interface {
	Get(string) (*models.WebauthnCredential, error)
	Create(models.WebauthnCredential) error
	Update(models.WebauthnCredential) error
	Delete(models.WebauthnCredential) error
	GetFromUser(uuid.UUID) ([]models.WebauthnCredential, error)
}

type webauthnCredentialPersister struct {
	db *pop.Connection
}

func NewWebauthnCredentialPersister(db *pop.Connection) WebauthnCredentialPersister {
	return &webauthnCredentialPersister{db: db}
}

func (p *webauthnCredentialPersister) Get(id string) (*models.WebauthnCredential, error) {
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

// Create stores a new `WebauthnCredential`. Please run inside a transaction, since `Transports` associated with the
// credential are stored separately in another table.
func (p *webauthnCredentialPersister) Create(credential models.WebauthnCredential) error {
	vErr, err := p.db.ValidateAndCreate(&credential)
	if err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("credential object validation failed: %w", vErr)
	}

	// Eager creation seems to be broken, so we need to store the transports separately.
	// See: https://github.com/gobuffalo/pop/issues/608
	vErr, err = p.db.ValidateAndCreate(&credential.Transports)
	if err != nil {
		return fmt.Errorf("failed to store credential transport: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("credential transport validation failed: %w", vErr)
	}

	return nil
}

func (p *webauthnCredentialPersister) Update(credential models.WebauthnCredential) error {
	vErr, err := p.db.ValidateAndUpdate(&credential)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("credential object validation failed: %w", vErr)
	}

	return nil
}

func (p *webauthnCredentialPersister) Delete(credential models.WebauthnCredential) error {
	err := p.db.Destroy(&credential)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	return nil
}

func (p *webauthnCredentialPersister) GetFromUser(userId uuid.UUID) ([]models.WebauthnCredential, error) {
	var credentials []models.WebauthnCredential
	err := p.db.Eager().Where("user_id = ?", &userId).All(&credentials)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return credentials, nil
}
