package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type UsernamePersister interface {
	Create(username models.Username) error
	GetByUserID(userId uuid.UUID) (*models.Username, error)
	GetByName(name string) (*models.Username, error)
	Update(username *models.Username) error
	Delete(username *models.Username) error
}

type usernamePersister struct {
	db *pop.Connection
}

func NewUsernamePersister(db *pop.Connection) UsernamePersister {
	return &usernamePersister{db: db}
}

func (p *usernamePersister) Create(username models.Username) error {
	vErr, err := p.db.ValidateAndCreate(&username)
	if err != nil {
		return fmt.Errorf("failed to store username credential: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("username object validation failed: %w", vErr)
	}

	return nil
}

func (p *usernamePersister) GetByUserID(userId uuid.UUID) (*models.Username, error) {
	pw := models.Username{}
	query := p.db.Where("user_id = (?)", userId.String())
	err := query.First(&pw)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	return &pw, nil
}

func (p *usernamePersister) GetByName(username string) (*models.Username, error) {
	pw := models.Username{}
	query := p.db.Where("username = (?)", username)
	err := query.First(&pw)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get username: %w", err)
	}
	return &pw, nil
}

func (p *usernamePersister) Update(username *models.Username) error {
	vErr, err := p.db.ValidateAndUpdate(username)
	if err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("username object validation failed: %w", vErr)
	}

	return nil
}

func (p *usernamePersister) Delete(username *models.Username) error {
	err := p.db.Destroy(username)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
