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
	Get(uuid.UUID) (*models.Username, error)
	Create(models.Username) error
	Update(models.Username) error
	Delete(models.Username) error
	Find(string) (*models.Username, error)
}

type usernamePersister struct {
	db *pop.Connection
}

func NewUsernamePersister(db *pop.Connection) UsernamePersister {
	return &usernamePersister{db: db}
}

func (p *usernamePersister) Get(id uuid.UUID) (*models.Username, error) {
	username := models.Username{}
	err := p.db.Find(&username, id.String())
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &username, nil
}

func (p *usernamePersister) Create(username models.Username) error {
	vErr, err := p.db.ValidateAndCreate(&username)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("username object validation failed: %w", vErr)
	}

	return nil
}

func (p *usernamePersister) Update(username models.Username) error {
	vErr, err := p.db.ValidateAndUpdate(&username)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("email object validation failed: %w", vErr)
	}

	return nil
}

func (p *usernamePersister) Delete(username models.Username) error {
	err := p.db.Destroy(&username)
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	return nil
}

func (p *usernamePersister) Find(username string) (*models.Username, error) {
	var u models.Username

	query := p.db.Where("username = ?", username)
	err := query.First(&u)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &u, nil
}
