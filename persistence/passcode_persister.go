package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/persistence/models"
)

type PasscodePersister struct {
	db *pop.Connection
}

func NewPasscodePersister(db *pop.Connection) *PasscodePersister {
	return &PasscodePersister{db: db}
}

func (p *PasscodePersister) Get(id uuid.UUID) (*models.Passcode, error) {
	passcode := models.Passcode{}
	err := p.db.Find(&passcode, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get passcode: %w", err)
	}

	return &passcode, nil
}

func (p *PasscodePersister) Create(passcode models.Passcode) error {
	vErr, err := p.db.ValidateAndCreate(&passcode)
	if err != nil {
		return fmt.Errorf("failed to store passcode: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("passcode object validation failed: %w", vErr)
	}

	return nil
}

func (p *PasscodePersister) Update(passcode models.Passcode) error {
	vErr, err := p.db.ValidateAndUpdate(&passcode)
	if err != nil {
		return fmt.Errorf("failed to update passcode: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("passcode object validation failed: %w", vErr)
	}

	return nil
}

func (p *PasscodePersister) Delete(passcode models.Passcode) error {
	err := p.db.Destroy(&passcode)
	if err != nil {
		return fmt.Errorf("failed to delete passcode: %w", err)
	}

	return nil
}
