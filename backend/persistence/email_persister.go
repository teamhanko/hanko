package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailPersister interface {
	FindByUserId(uuid.UUID) ([]models.Email, error)
	FindByAddress(string) (*models.Email, error)
	Create(models.Email) error
	Update(models.Email) error
	Delete(models.Email) error
}

type emailPersister struct {
	db *pop.Connection
}

func NewEmailPersister(db *pop.Connection) EmailPersister {
	return &emailPersister{db: db}
}

func (e *emailPersister) FindByUserId(userId uuid.UUID) ([]models.Email, error) {
	var emails []models.Email

	err := e.db.Where("user_id = ?", userId.String()).Order("created_at desc").All(&emails)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return emails, nil
	}

	if err != nil {
		return nil, err
	}

	return emails, nil
}

func (e *emailPersister) FindByAddress(address string) (*models.Email, error) {
	var email models.Email

	query := e.db.EagerPreload().Where("address = ?", address)
	err := query.First(&email)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &email, nil
}

func (e *emailPersister) Create(email models.Email) error {
	vErr, err := e.db.ValidateAndCreate(&email)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("email object validation failed: %w", vErr)
	}

	return nil
}

func (e *emailPersister) Update(email models.Email) error {
	vErr, err := e.db.ValidateAndUpdate(&email)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("email object validation failed: %w", vErr)
	}

	return nil
}

func (e *emailPersister) Delete(email models.Email) error {
	err := e.db.Eager().Destroy(&email)
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	return nil
}
