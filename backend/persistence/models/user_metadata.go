package models

import (
	"fmt"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type UserMetadata struct {
	ID        uuid.UUID    `db:"id"`
	UserID    uuid.UUID    `db:"user_id"`
	Public    nulls.String `db:"public_metadata"`
	Private   nulls.String `db:"private_metadata"`
	Unsafe    nulls.String `db:"unsafe_metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

func (m *UserMetadata) Validate(tx *pop.Connection) (*validate.Errors, error) {
	metadataMax := 3000

	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: m.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: m.UserID},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: m.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: m.CreatedAt},
		&validators.StringLengthInRange{
			Name:    "Public",
			Field:   m.Public.String,
			Max:     metadataMax,
			Message: fmt.Sprintf("public metadata must not exceed %d characters", metadataMax),
		},
		&validators.StringLengthInRange{
			Name:    "Private",
			Field:   m.Private.String,
			Max:     metadataMax,
			Message: fmt.Sprintf("private metadata must not exceed %d characters", metadataMax),
		},
		&validators.StringLengthInRange{
			Name:    "Unsafe",
			Field:   m.Unsafe.String,
			Max:     metadataMax,
			Message: fmt.Sprintf("unsafe metadata must not exceed %d characters", metadataMax),
		},
	), nil
}
