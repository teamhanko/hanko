package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type UsernamePersister interface {
	Create(username models.Username) error
	GetByName(name string) (*models.Username, error)
	GetByNameAndTenant(name string, tenantID *uuid.UUID) (*models.Username, error)
	// GetByNameWithTenantFallback looks for username:
	// 1. First tries to find with the specified tenant_id
	// 2. If not found and tenantID is provided, falls back to global (tenant_id IS NULL)
	// Returns: username, isGlobalFallback, error
	GetByNameWithTenantFallback(name string, tenantID *uuid.UUID) (*models.Username, bool, error)
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

func (p *usernamePersister) GetByNameAndTenant(username string, tenantID *uuid.UUID) (*models.Username, error) {
	pw := models.Username{}
	query := p.db.Where("username = (?)", username)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID.String())
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.First(&pw)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get username: %w", err)
	}
	return &pw, nil
}

func (p *usernamePersister) GetByNameWithTenantFallback(username string, tenantID *uuid.UUID) (*models.Username, bool, error) {
	// If no tenant specified, just do normal lookup
	if tenantID == nil {
		un, err := p.GetByName(username)
		return un, false, err
	}

	// First, try to find with the specified tenant
	un, err := p.GetByNameAndTenant(username, tenantID)
	if err != nil {
		return nil, false, err
	}
	if un != nil {
		return un, false, nil // Found in tenant
	}

	// Not found in tenant, try to find global user (tenant_id IS NULL)
	un, err = p.GetByNameAndTenant(username, nil)
	if err != nil {
		return nil, false, err
	}
	if un != nil {
		return un, true, nil // Found as global user (needs adoption)
	}

	return nil, false, nil // Not found anywhere
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
