package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// RoleBinding represents a role held by a user within one organization,
// including the role's own slug/name so a caller doesn't need a second
// GET /roles/{id} to display it.
type RoleBinding struct {
	RoleID    uuid.UUID `json:"role_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// FromRoleBindingModel converts a role binding and its role to a DTO object
func FromRoleBindingModel(binding models.RoleBinding, role models.Role) RoleBinding {
	return RoleBinding{
		RoleID:    role.ID,
		Slug:      role.Slug,
		Name:      role.Name,
		CreatedAt: binding.CreatedAt,
	}
}

// CreateRoleBindingRequest.Role may reference a role by either its id or
// its slug - resolved via RolePersister.GetByIDOrSlug.
type CreateRoleBindingRequest struct {
	Role string `json:"role" validate:"required"`
}
