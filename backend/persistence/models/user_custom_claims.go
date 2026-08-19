package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// UserCustomClaims holds the resolved, tenant-declared custom claims for one user - see
// config.CustomClaims and thirdparty.ResolveCustomClaims. Claims is a JSON object keyed by
// claim name, each entry shaped like:
//
//	{"matriculation_number": {"value": "12345", "source": "saml:<provider_id>"}}
//
// source is the writing connection's identifier ("saml:<provider_id>", "oidc:<provider_id>"),
// or "admin" for a value set via the Admin API. It stays internal to this table.
type UserCustomClaims struct {
	ID        uuid.UUID       `db:"id"`
	UserID    uuid.UUID       `db:"user_id"`
	TenantID  uuid.UUID       `db:"tenant_id"`
	Claims    json.RawMessage `db:"claims"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}

func (c *UserCustomClaims) Validate(_ *pop.Connection) (*validate.Errors, error) {
	claimsMax := 3000

	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: c.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: c.UserID},
		&validators.UUIDIsPresent{Name: "TenantID", Field: c.TenantID},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: c.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: c.CreatedAt},
		&validators.StringLengthInRange{
			Name:    "Claims",
			Field:   string(c.Claims),
			Max:     claimsMax,
			Message: fmt.Sprintf("custom claims must not exceed %d characters", claimsMax),
		},
	), nil
}
