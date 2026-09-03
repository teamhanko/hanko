package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// OrganizationMembership is used by pop to map your organization_memberships database table to your go code.
// It records a user belonging to an organization, independent of any role.
type OrganizationMembership struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TenantID       uuid.UUID `json:"-" db:"tenant_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type OrganizationMemberships []OrganizationMembership

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (m *OrganizationMembership) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: m.ID},
		&validators.UUIDIsPresent{Name: "TenantID", Field: m.TenantID},
		&validators.UUIDIsPresent{Name: "UserID", Field: m.UserID},
		&validators.UUIDIsPresent{Name: "OrganizationID", Field: m.OrganizationID},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: m.CreatedAt},
	), nil
}
