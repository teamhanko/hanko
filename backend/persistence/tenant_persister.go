package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type TenantPersister interface {
	Get(id uuid.UUID) (*models.Tenant, error)
	Create(tenant models.Tenant) error
	Update(tenant models.Tenant) error
	Delete(tenant models.Tenant) error
	List(page int, perPage int) ([]models.Tenant, error)
	All() ([]models.Tenant, error)
	Count() (int, error)
}

type tenantPersister struct {
	db *pop.Connection
}

func NewTenantPersister(db *pop.Connection) TenantPersister {
	return &tenantPersister{db: db}
}

func (p *tenantPersister) Get(id uuid.UUID) (*models.Tenant, error) {
	tenant := models.Tenant{}
	err := p.db.Find(&tenant, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

func (p *tenantPersister) Create(tenant models.Tenant) error {
	vErr, err := p.db.ValidateAndCreate(&tenant)
	if err != nil {
		return fmt.Errorf("failed to store tenant: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("tenant object validation failed: %w", vErr)
	}

	return nil
}

func (p *tenantPersister) Update(tenant models.Tenant) error {
	vErr, err := p.db.ValidateAndUpdate(&tenant)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("tenant object validation failed: %w", vErr)
	}

	return nil
}

func (p *tenantPersister) Delete(tenant models.Tenant) error {
	err := p.db.Destroy(&tenant)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

func (p *tenantPersister) List(page int, perPage int) ([]models.Tenant, error) {
	tenants := []models.Tenant{}

	err := p.db.
		Order("created_at desc").
		Paginate(page, perPage).
		All(&tenants)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return tenants, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenants: %w", err)
	}

	return tenants, nil
}

func (p *tenantPersister) All() ([]models.Tenant, error) {
	tenants := []models.Tenant{}

	err := p.db.All(&tenants)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return tenants, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenants: %w", err)
	}

	return tenants, nil
}

func (p *tenantPersister) Count() (int, error) {
	count, err := p.db.Count(&models.Tenant{})
	if err != nil {
		return 0, fmt.Errorf("failed to get tenant count: %w", err)
	}

	return count, nil
}
