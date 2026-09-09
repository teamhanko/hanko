package models

import "github.com/gofrs/uuid"

// UserOrganizationRole is a role a user holds within one of their organizations.
type UserOrganizationRole struct {
	ID   uuid.UUID `json:"id"`
	Slug string    `json:"slug"`
	Name string    `json:"name"`
}

// UserOrganizationRoles is a query-result shape, not a persisted table - it
// combines one organization a user belongs to with the roles they hold in
// it. Produced by
// OrganizationMembershipPersister.ListOrganizationsWithRolesByUserIDs.
type UserOrganizationRoles struct {
	OrganizationID   uuid.UUID              `json:"organization_id"`
	OrganizationName string                 `json:"organization_name"`
	Roles            []UserOrganizationRole `json:"roles"`
}
