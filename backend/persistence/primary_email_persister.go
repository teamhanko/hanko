package persistence

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PrimaryEmailPersister interface {
	Create(models.PrimaryEmail) error
	Update(models.PrimaryEmail) error
}

type primaryEmailPersister struct {
	db *pop.Connection
}

func NewPrimaryEmailPersister(db *pop.Connection) PrimaryEmailPersister {
	return &primaryEmailPersister{db: db}
}

func (p *primaryEmailPersister) Create(primaryEmail models.PrimaryEmail) error {
	vErr, err := p.db.ValidateAndCreate(&primaryEmail)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("primary email object validation failed: %w", vErr)
	}

	return nil
}

func (p *primaryEmailPersister) Update(primaryEmail models.PrimaryEmail) error {
	vErr, err := p.db.ValidateAndSave(&primaryEmail)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("primary email object validation failed: %w", vErr)
	}

	return nil
}
