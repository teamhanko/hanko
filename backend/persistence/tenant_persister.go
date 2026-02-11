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
	GetBySlug(slug string) (*models.Tenant, error)
	Create(tenant models.Tenant) error
	Update(tenant models.Tenant) error
	Delete(tenant models.Tenant) error
	List(page, perPage int) ([]models.Tenant, error)
	Count() (int, error)
}

type tenantPersister struct {
	db *pop.Connection
}

func NewTenantPersister(db *pop.Connection) TenantPersister {
	return &tenantPersister{db: db}
}

func (t *tenantPersister) Get(id uuid.UUID) (*models.Tenant, error) {
	tenant := models.Tenant{}
	err := t.db.Find(&tenant, id.String())
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return &tenant, nil
}

func (t *tenantPersister) GetBySlug(slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := t.db.Where("slug = ?", slug).First(&tenant)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by slug: %w", err)
	}
	return &tenant, nil
}

func (t *tenantPersister) Create(tenant models.Tenant) error {
	vErr, err := t.db.ValidateAndCreate(&tenant)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("tenant validation failed: %w", vErr)
	}
	return nil
}

func (t *tenantPersister) Update(tenant models.Tenant) error {
	vErr, err := t.db.ValidateAndUpdate(&tenant)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("tenant validation failed: %w", vErr)
	}
	return nil
}

func (t *tenantPersister) Delete(tenant models.Tenant) error {
	err := t.db.Destroy(&tenant)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

func (t *tenantPersister) List(page, perPage int) ([]models.Tenant, error) {
	var tenants []models.Tenant
	query := t.db.Order("created_at desc")
	if page > 0 && perPage > 0 {
		query = query.Paginate(page, perPage)
	}
	err := query.All(&tenants)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return tenants, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

func (t *tenantPersister) Count() (int, error) {
	var tenants []models.Tenant
	count, err := t.db.Count(&tenants)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}
	return count, nil
}
