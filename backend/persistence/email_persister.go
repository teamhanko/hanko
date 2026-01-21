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
	Get(emailId uuid.UUID) (*models.Email, error)
	CountByUserId(uuid.UUID) (int, error)
	FindByUserId(uuid.UUID) (models.Emails, error)
	FindByAddress(string) (*models.Email, error)
	FindByAddressAndTenant(address string, tenantID *uuid.UUID) (*models.Email, error)
	// FindByAddressWithTenantFallback looks for email by address:
	// 1. First tries to find with the specified tenant_id
	// 2. If not found and tenantID is provided, falls back to global (tenant_id IS NULL)
	// Returns: email, isGlobalFallback, error
	FindByAddressWithTenantFallback(address string, tenantID *uuid.UUID) (*models.Email, bool, error)
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

func (e *emailPersister) Get(emailId uuid.UUID) (*models.Email, error) {
	email := models.Email{}
	err := e.db.Find(&email, emailId.String())
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &email, nil
}

func (e *emailPersister) FindByUserId(userId uuid.UUID) (models.Emails, error) {
	var emails models.Emails

	err := e.db.EagerPreload().
		Where("user_id = ?", userId.String()).
		Order("created_at asc").
		All(&emails)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return emails, nil
	}

	if err != nil {
		return nil, err
	}

	return emails, nil
}

func (e *emailPersister) CountByUserId(userId uuid.UUID) (int, error) {
	var emails []models.Email

	count, err := e.db.
		Where("user_id = ?", userId.String()).
		Count(&emails)

	if err != nil {
		return 0, err
	}

	return count, nil
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

func (e *emailPersister) FindByAddressAndTenant(address string, tenantID *uuid.UUID) (*models.Email, error) {
	var email models.Email

	query := e.db.EagerPreload().Where("address = ?", address)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID.String())
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

func (e *emailPersister) FindByAddressWithTenantFallback(address string, tenantID *uuid.UUID) (*models.Email, bool, error) {
	// If no tenant specified, just do normal lookup
	if tenantID == nil {
		email, err := e.FindByAddress(address)
		return email, false, err
	}

	// First, try to find with the specified tenant
	email, err := e.FindByAddressAndTenant(address, tenantID)
	if err != nil {
		return nil, false, err
	}
	if email != nil {
		return email, false, nil // Found in tenant
	}

	// Not found in tenant, try to find global user (tenant_id IS NULL)
	email, err = e.FindByAddressAndTenant(address, nil)
	if err != nil {
		return nil, false, err
	}
	if email != nil {
		return email, true, nil // Found as global user (needs adoption)
	}

	return nil, false, nil // Not found anywhere
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
