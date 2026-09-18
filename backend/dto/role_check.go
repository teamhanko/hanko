package dto

import "github.com/gofrs/uuid"

// RoleCheckRequest.Roles entries may reference a role by either its id or
// its slug - resolved via RolePersister.GetByIDOrSlug.
type RoleCheckRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	Roles          []string  `json:"roles" validate:"required,min=1"`
}

type RoleCheckResponse struct {
	HasRole bool `json:"has_role"`
}
