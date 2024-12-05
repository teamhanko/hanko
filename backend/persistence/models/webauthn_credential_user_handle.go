package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type WebauthnCredentialUserHandle struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Handle    string    `db:"handle" json:"handle"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (userHandle *WebauthnCredentialUserHandle) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: userHandle.ID},
		&validators.UUIDIsPresent{Name: "UserId", Field: userHandle.UserID},
		&validators.StringIsPresent{Name: "handle", Field: userHandle.Handle},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: userHandle.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: userHandle.UpdatedAt},
	), nil
}
