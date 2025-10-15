package persistence

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type SamlIdentityPersister interface {
	Create(samlIdentity models.SamlIdentity) error
	Update(samlIdentity models.SamlIdentity) error
}

type samlIdentityPersister struct {
	db *pop.Connection
}

func NewSamlIdentityPersister(db *pop.Connection) SamlIdentityPersister {
	return &samlIdentityPersister{db: db}
}

func (p samlIdentityPersister) Create(samlIdentity models.SamlIdentity) error {
	vErr, err := p.db.Eager().ValidateAndCreate(&samlIdentity)
	if err != nil {
		return fmt.Errorf("failed to store saml identity: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("saml identity object validation failed: %w", vErr)
	}

	return nil
}

func (p samlIdentityPersister) Update(samlIdentity models.SamlIdentity) error {
	vErr, err := p.db.ValidateAndUpdate(&samlIdentity)
	if err != nil {
		return fmt.Errorf("failed to update saml identity: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("saml identity object validation failed: %w", vErr)
	}

	return nil
}
