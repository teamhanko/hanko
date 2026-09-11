package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// RoleBinding is used by pop to map your role_bindings database table to your go code.
// It records a role held by a user within one specific organization.
type RoleBinding struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TenantID       uuid.UUID `json:"-" db:"tenant_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	RoleID         uuid.UUID `json:"role_id" db:"role_id"`
	Role           *Role     `json:"role,omitempty" belongs_to:"role"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type RoleBindings []RoleBinding

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (b *RoleBinding) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: b.ID},
		&validators.UUIDIsPresent{Name: "TenantID", Field: b.TenantID},
		&validators.UUIDIsPresent{Name: "UserID", Field: b.UserID},
		&validators.UUIDIsPresent{Name: "RoleID", Field: b.RoleID},
		&validators.UUIDIsPresent{Name: "OrganizationID", Field: b.OrganizationID},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: b.CreatedAt},
	), nil
}
