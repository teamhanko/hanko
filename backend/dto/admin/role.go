package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type Role struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromRoleModel converts the DB model to a DTO object
func FromRoleModel(model models.Role) Role {
	return Role{
		ID:        model.ID,
		Slug:      model.Slug,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

type CreateRole struct {
	Slug string `json:"slug" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// PatchRoleRequest only allows updating the display name. Slug is
// immutable after creation - it's referenced by role-binding requests
// ({"role": "<id or slug>"}) and by relying parties in the public
// role-check endpoint, so letting it change would silently break any
// in-flight or hardcoded reference.
type PatchRoleRequest struct {
	Name dto.OptionalString `json:"name"`
}
