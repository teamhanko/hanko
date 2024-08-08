package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasslinkPersister interface {
	Get(uuid.UUID) (*models.Passlink, error)
	Create(models.Passlink) error
	Update(models.Passlink) error
	Delete(models.Passlink) error
}

type passlinkPersister struct {
	db *pop.Connection
}

func NewPasslinkPersister(db *pop.Connection) PasslinkPersister {
	return &passlinkPersister{db: db}
}

func (p *passlinkPersister) Get(id uuid.UUID) (*models.Passlink, error) {
	passlink := models.Passlink{}
	err := p.db.EagerPreload("Email.User").Find(&passlink, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get passlink: %w", err)
	}

	return &passlink, nil
}

func (p *passlinkPersister) Create(passlink models.Passlink) error {
	vErr, err := p.db.ValidateAndCreate(&passlink)
	if err != nil {
		return fmt.Errorf("failed to store passlink: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("passlink object validation failed: %w", vErr)
	}

	return nil
}

func (p *passlinkPersister) Update(passlink models.Passlink) error {
	vErr, err := p.db.ValidateAndUpdate(&passlink)
	if err != nil {
		return fmt.Errorf("failed to update passlink: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("passlink object validation failed: %w", vErr)
	}

	return nil
}

func (p *passlinkPersister) Delete(passlink models.Passlink) error {
	err := p.db.Destroy(&passlink)
	if err != nil {
		return fmt.Errorf("failed to delete passlink: %w", err)
	}

	return nil
}
