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
	// ListOrganizationsWithRolesByUserIDs returns, for every user in
	// userIDs, the organizations they belong to and the roles they hold in
	// each - one query for the whole batch, not one per user.
	ListOrganizationsWithRolesByUserIDs(userIDs []uuid.UUID, tenantID uuid.UUID) (map[uuid.UUID][]models.UserOrganizationRoles, error)
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

type organizationWithRoleRow struct {
	UserID           uuid.UUID      `db:"user_id"`
	OrganizationID   uuid.UUID      `db:"organization_id"`
	OrganizationName string         `db:"organization_name"`
	RoleID           uuid.NullUUID  `db:"role_id"`
	RoleSlug         sql.NullString `db:"role_slug"`
	RoleName         sql.NullString `db:"role_name"`
}

type userOrgKey struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
}

// ListOrganizationsWithRolesByUserIDs is a hand-written join rather than
// EagerPreload, so that batching multiple users costs one query total
// instead of one per user: pop's association batching always costs one
// query per named relationship, which is fine for a single user but would
// mean N users on a page each paying that same cost again.
func (p *organizationMembershipPersister) ListOrganizationsWithRolesByUserIDs(userIDs []uuid.UUID, tenantID uuid.UUID) (map[uuid.UUID][]models.UserOrganizationRoles, error) {
	return listOrganizationsWithRolesByUserIDs(p.db, userIDs, tenantID)
}

// listOrganizationsWithRolesByUserIDs is shared by
// organizationMembershipPersister and userPersister (Get/List/GetByUsername),
// so the latter doesn't need a reference to the former - every persister
// already only holds a *pop.Connection.
func listOrganizationsWithRolesByUserIDs(db *pop.Connection, userIDs []uuid.UUID, tenantID uuid.UUID) (map[uuid.UUID][]models.UserOrganizationRoles, error) {
	result := make(map[uuid.UUID][]models.UserOrganizationRoles)
	if len(userIDs) == 0 {
		return result, nil
	}

	rows := []organizationWithRoleRow{}
	err := db.RawQuery(`
		SELECT om.user_id AS user_id,
		       om.organization_id AS organization_id,
		       o.name AS organization_name,
		       rb.role_id AS role_id,
		       r.slug AS role_slug,
		       r.name AS role_name
		FROM organization_memberships om
		JOIN organizations o ON o.id = om.organization_id AND o.tenant_id = om.tenant_id
		LEFT JOIN role_bindings rb ON rb.user_id = om.user_id AND rb.organization_id = om.organization_id AND rb.tenant_id = om.tenant_id
		LEFT JOIN roles r ON r.id = rb.role_id AND r.tenant_id = rb.tenant_id
		WHERE om.user_id IN (?) AND om.tenant_id = ?
		ORDER BY om.created_at, rb.created_at
	`, userIDs, tenantID).All(&rows)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizations with roles: %w", err)
	}

	orgIndex := make(map[userOrgKey]int)
	for _, row := range rows {
		key := userOrgKey{UserID: row.UserID, OrganizationID: row.OrganizationID}
		idx, ok := orgIndex[key]
		if !ok {
			result[row.UserID] = append(result[row.UserID], models.UserOrganizationRoles{
				OrganizationID:   row.OrganizationID,
				OrganizationName: row.OrganizationName,
			})
			idx = len(result[row.UserID]) - 1
			orgIndex[key] = idx
		}

		if row.RoleID.Valid {
			result[row.UserID][idx].Roles = append(result[row.UserID][idx].Roles, models.UserOrganizationRole{
				ID:   row.RoleID.UUID,
				Slug: row.RoleSlug.String,
				Name: row.RoleName.String,
			})
		}
	}

	return result, nil
}
