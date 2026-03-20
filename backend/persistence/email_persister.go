package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type EmailPersister interface {
	Get(emailId uuid.UUID, tenantID *uuid.UUID) (*models.Email, error)
	CountByUserId(userId uuid.UUID, tenantID *uuid.UUID) (int, error)
	FindByUserId(userId uuid.UUID, tenantID *uuid.UUID) (models.Emails, error)
	FindByAddress(address string, tenantID *uuid.UUID) (*models.Email, error)
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

func (e *emailPersister) Get(emailId uuid.UUID, tenantID *uuid.UUID) (*models.Email, error) {
	email := models.Email{}
	query := e.db.Q()
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.Find(&email, emailId.String())
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &email, nil
}

func (e *emailPersister) FindByUserId(userId uuid.UUID, tenantID *uuid.UUID) (models.Emails, error) {
	var emails models.Emails

	query := e.db.EagerPreload().Where("user_id = ?", userId.String())
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.Order("created_at asc").All(&emails)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return emails, nil
	}

	if err != nil {
		return nil, err
	}

	return emails, nil
}

func (e *emailPersister) CountByUserId(userId uuid.UUID, tenantID *uuid.UUID) (int, error) {
	var emails []models.Email

	query := e.db.Where("user_id = ?", userId.String())
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	count, err := query.Count(&emails)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (e *emailPersister) FindByAddress(address string, tenantID *uuid.UUID) (*models.Email, error) {
	var email models.Email

	query := e.db.EagerPreload().Where("address = ?", address)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
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
	err := e.db.Destroy(&email)
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	return nil
}
