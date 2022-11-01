package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type PrimaryEmail struct {
	ID        uuid.UUID `db:"id" json:"id"`
	EmailID   uuid.UUID `db:"email_id" json:"email_id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Email     *Email    `belongs_to:"email" json:"email,omitempty"`
	User      *User     `belongs_to:"user" json:"user,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func NewPrimaryEmail(emailId uuid.UUID, userId uuid.UUID) *PrimaryEmail {
	id, _ := uuid.NewV4()

	return &PrimaryEmail{
		ID:        id,
		EmailID:   emailId,
		UserID:    userId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (primaryEmail *PrimaryEmail) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: primaryEmail.ID},
		&validators.UUIDIsPresent{Name: "EmailID", Field: primaryEmail.EmailID},
		&validators.UUIDIsPresent{Name: "UserID", Field: primaryEmail.UserID},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: primaryEmail.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: primaryEmail.CreatedAt},
	), nil
}
