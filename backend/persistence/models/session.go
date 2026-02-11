package models

import (
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type Session struct {
	ID        uuid.UUID    `db:"id" json:"id"`
	TenantID  *uuid.UUID   `db:"tenant_id" json:"tenant_id,omitempty"`
	UserID    uuid.UUID    `db:"user_id" json:"user_id"`
	UserAgent nulls.String `db:"user_agent" json:"user_agent"`
	IpAddress nulls.String `db:"ip_address" json:"ip_address"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	ExpiresAt *time.Time   `db:"expires_at" json:"expires_at"`
	LastUsed  time.Time    `db:"last_used" json:"last_used"`
}

func (session *Session) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: session.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: session.UserID},
		&validators.TimeIsPresent{Name: "LastUsed", Field: session.UpdatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: session.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: session.CreatedAt},
	), nil
}
