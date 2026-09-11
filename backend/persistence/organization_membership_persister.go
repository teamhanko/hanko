package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type OrganizationMembershipPersister interface {
	Get(userID uuid.UUID, organizationID uuid.UUID, tenantID uuid.UUID) (*models.OrganizationMembership, error)
	Create(membership models.OrganizationMembership) error
	Delete(membership models.OrganizationMembership) error
	ListByOrganization(organizationID uuid.UUID, page int, perPage int, tenantID uuid.UUID) ([]models.OrganizationMembership, error)
	CountByOrganization(organizationID uuid.UUID, tenantID uuid.UUID) (int, error)
	// ListByUser returns every organization userID belongs to. Not
	// paginated - used internally by /sessions/validate to compute the
	// full organizations claim on every call, not exposed via an API list.
	ListByUser(userID uuid.UUID, tenantID uuid.UUID) ([]models.OrganizationMembership, error)
}

type organizationMembershipPersister struct {
	db *pop.Connection
}

func NewOrganizationMembershipPersister(db *pop.Connection) OrganizationMembershipPersister {
	return &organizationMembershipPersister{db: db}
}

func (p *organizationMembershipPersister) Get(userID uuid.UUID, organizationID uuid.UUID, tenantID uuid.UUID) (*models.OrganizationMembership, error) {
	membership := models.OrganizationMembership{}
	err := p.db.
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		Where("organization_id = ?", organizationID).
		First(&membership)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization membership: %w", err)
	}

	return &membership, nil
}

func (p *organizationMembershipPersister) Create(membership models.OrganizationMembership) error {
	vErr, err := p.db.ValidateAndCreate(&membership)
	if err != nil {
		return fmt.Errorf("failed to store organization membership: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("organization membership object validation failed: %w", vErr)
	}

	return nil
}

func (p *organizationMembershipPersister) Delete(membership models.OrganizationMembership) error {
	err := p.db.Destroy(&membership)
	if err != nil {
		return fmt.Errorf("failed to delete organization membership: %w", err)
	}

	return nil
}

func (p *organizationMembershipPersister) ListByOrganization(organizationID uuid.UUID, page int, perPage int, tenantID uuid.UUID) ([]models.OrganizationMembership, error) {
	memberships := []models.OrganizationMembership{}

	err := p.db.
		Where("tenant_id = ?", tenantID).
		Where("organization_id = ?", organizationID).
		Order("created_at desc").
		Paginate(page, perPage).
		All(&memberships)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return memberships, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization memberships: %w", err)
	}

	return memberships, nil
}

func (p *organizationMembershipPersister) CountByOrganization(organizationID uuid.UUID, tenantID uuid.UUID) (int, error) {
	count, err := p.db.
		Where("tenant_id = ?", tenantID).
		Where("organization_id = ?", organizationID).
		Count(&models.OrganizationMembership{})
	if err != nil {
		return 0, fmt.Errorf("failed to get organization membership count: %w", err)
	}

	return count, nil
}

// ListByUser eager-loads the Organization association so callers get the
// organization name without an extra Get() call per membership - pop
// batches this into one additional query (an IN on organization ids)
// rather than one per row.
func (p *organizationMembershipPersister) ListByUser(userID uuid.UUID, tenantID uuid.UUID) ([]models.OrganizationMembership, error) {
	memberships := []models.OrganizationMembership{}

	err := p.db.
		EagerPreload("Organization").
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		Order("created_at desc").
		All(&memberships)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return memberships, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization memberships: %w", err)
	}

	return memberships, nil
}
