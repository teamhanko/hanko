package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type RolePersister interface {
	Get(id uuid.UUID, tenantID uuid.UUID) (*models.Role, error)
	GetBySlug(slug string, tenantID uuid.UUID) (*models.Role, error)
	// GetByIDOrSlug resolves ref as a UUID first; if ref does not parse as a
	// UUID, it is looked up as a slug. Not-found returns (nil, nil) in both cases.
	GetByIDOrSlug(ref string, tenantID uuid.UUID) (*models.Role, error)
	Create(role models.Role) error
	Update(role models.Role) error
	Delete(role models.Role) error
	List(page int, perPage int, tenantID uuid.UUID) ([]models.Role, error)
	Count(tenantID uuid.UUID) (int, error)
}

type rolePersister struct {
	db *pop.Connection
}

func NewRolePersister(db *pop.Connection) RolePersister {
	return &rolePersister{db: db}
}

func (p *rolePersister) Get(id uuid.UUID, tenantID uuid.UUID) (*models.Role, error) {
	role := models.Role{}
	err := p.db.Where("roles.tenant_id = ?", tenantID).Find(&role, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (p *rolePersister) GetBySlug(slug string, tenantID uuid.UUID) (*models.Role, error) {
	role := models.Role{}
	err := p.db.Where("tenant_id = ?", tenantID).Where("slug = ?", slug).First(&role)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role by slug: %w", err)
	}

	return &role, nil
}

func (p *rolePersister) GetByIDOrSlug(ref string, tenantID uuid.UUID) (*models.Role, error) {
	if id, err := uuid.FromString(ref); err == nil {
		return p.Get(id, tenantID)
	}

	return p.GetBySlug(ref, tenantID)
}

func (p *rolePersister) Create(role models.Role) error {
	vErr, err := p.db.ValidateAndCreate(&role)
	if err != nil {
		return fmt.Errorf("failed to store role: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("role object validation failed: %w", vErr)
	}

	return nil
}

func (p *rolePersister) Update(role models.Role) error {
	vErr, err := p.db.ValidateAndUpdate(&role)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("role object validation failed: %w", vErr)
	}

	return nil
}

func (p *rolePersister) Delete(role models.Role) error {
	err := p.db.Destroy(&role)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

func (p *rolePersister) List(page int, perPage int, tenantID uuid.UUID) ([]models.Role, error) {
	roles := []models.Role{}

	err := p.db.
		Where("tenant_id = ?", tenantID).
		Order("created_at desc").
		Paginate(page, perPage).
		All(&roles)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return roles, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	return roles, nil
}

func (p *rolePersister) Count(tenantID uuid.UUID) (int, error) {
	count, err := p.db.Where("tenant_id = ?", tenantID).Count(&models.Role{})
	if err != nil {
		return 0, fmt.Errorf("failed to get role count: %w", err)
	}

	return count, nil
}
