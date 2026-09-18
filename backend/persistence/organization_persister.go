package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type OrganizationPersister interface {
	Get(id uuid.UUID, tenantID uuid.UUID) (*models.Organization, error)
	GetByName(name string, tenantID uuid.UUID) (*models.Organization, error)
	Create(organization models.Organization) error
	Update(organization models.Organization) error
	Delete(organization models.Organization) error
	List(page int, perPage int, tenantID uuid.UUID) ([]models.Organization, error)
	Count(tenantID uuid.UUID) (int, error)
}

type organizationPersister struct {
	db *pop.Connection
}

func NewOrganizationPersister(db *pop.Connection) OrganizationPersister {
	return &organizationPersister{db: db}
}

func (p *organizationPersister) Get(id uuid.UUID, tenantID uuid.UUID) (*models.Organization, error) {
	organization := models.Organization{}
	err := p.db.Where("organizations.tenant_id = ?", tenantID).Find(&organization, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &organization, nil
}

func (p *organizationPersister) GetByName(name string, tenantID uuid.UUID) (*models.Organization, error) {
	organization := models.Organization{}
	err := p.db.Where("tenant_id = ?", tenantID).Where("name = ?", name).First(&organization)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization by name: %w", err)
	}

	return &organization, nil
}

func (p *organizationPersister) Create(organization models.Organization) error {
	vErr, err := p.db.ValidateAndCreate(&organization)
	if err != nil {
		return fmt.Errorf("failed to store organization: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("organization object validation failed: %w", vErr)
	}

	return nil
}

func (p *organizationPersister) Update(organization models.Organization) error {
	vErr, err := p.db.ValidateAndUpdate(&organization)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("organization object validation failed: %w", vErr)
	}

	return nil
}

func (p *organizationPersister) Delete(organization models.Organization) error {
	err := p.db.Destroy(&organization)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

func (p *organizationPersister) List(page int, perPage int, tenantID uuid.UUID) ([]models.Organization, error) {
	organizations := []models.Organization{}

	err := p.db.
		Where("tenant_id = ?", tenantID).
		Order("created_at desc").
		Paginate(page, perPage).
		All(&organizations)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return organizations, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizations: %w", err)
	}

	return organizations, nil
}

func (p *organizationPersister) Count(tenantID uuid.UUID) (int, error) {
	count, err := p.db.Where("tenant_id = ?", tenantID).Count(&models.Organization{})
	if err != nil {
		return 0, fmt.Errorf("failed to get organization count: %w", err)
	}

	return count, nil
}
