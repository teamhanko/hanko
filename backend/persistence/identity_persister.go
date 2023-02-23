package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type IdentityPersister interface {
	Get(userProviderID string, providerID string) (*models.Identity, error)
	Create(identity models.Identity) error
	Update(identity models.Identity) error
	Delete(identity models.Identity) error
}

type identityPersister struct {
	db *pop.Connection
}

func (p identityPersister) Get(userProviderID string, providerID string) (*models.Identity, error) {
	identity := &models.Identity{}
	if err := p.db.EagerPreload().Where("provider_id = ? AND provider_name = ?", userProviderID, providerID).First(identity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	return identity, nil
}

func (p identityPersister) Create(identity models.Identity) error {
	vErr, err := p.db.ValidateAndCreate(&identity)
	if err != nil {
		return fmt.Errorf("failed to store identity: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("identity object validation failed: %w", vErr)
	}

	return nil
}

func (p identityPersister) Update(identity models.Identity) error {
	vErr, err := p.db.ValidateAndUpdate(&identity)
	if err != nil {
		return fmt.Errorf("failed to update identity: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("identity object validation failed: %w", vErr)
	}

	return nil
}

func (p identityPersister) Delete(identity models.Identity) error {
	err := p.db.Destroy(&identity)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	return nil
}

func NewIdentityPersister(db *pop.Connection) IdentityPersister {
	return &identityPersister{db: db}
}
