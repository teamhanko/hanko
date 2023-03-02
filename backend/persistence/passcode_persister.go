package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasscodePersister interface {
	Get(uuid.UUID) (*models.Passcode, error)
	Create(models.Passcode) error
	Update(models.Passcode) error
	Delete(models.Passcode) error
}

type passcodePersister struct {
	db *pop.Connection
}

func NewPasscodePersister(db *pop.Connection) PasscodePersister {
	return &passcodePersister{db: db}
}

func (p *passcodePersister) Get(id uuid.UUID) (*models.Passcode, error) {
	passcode := models.Passcode{}
	err := p.db.EagerPreload("Email.User").Find(&passcode, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get passcode: %w", err)
	}

	return &passcode, nil
}

func (p *passcodePersister) Create(passcode models.Passcode) error {
	vErr, err := p.db.ValidateAndCreate(&passcode)
	if err != nil {
		return fmt.Errorf("failed to store passcode: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("passcode object validation failed: %w", vErr)
	}

	return nil
}

func (p *passcodePersister) Update(passcode models.Passcode) error {
	vErr, err := p.db.ValidateAndUpdate(&passcode)
	if err != nil {
		return fmt.Errorf("failed to update passcode: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("passcode object validation failed: %w", vErr)
	}

	return nil
}

func (p *passcodePersister) Delete(passcode models.Passcode) error {
	err := p.db.Destroy(&passcode)
	if err != nil {
		return fmt.Errorf("failed to delete passcode: %w", err)
	}

	return nil
}
