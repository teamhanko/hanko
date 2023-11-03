package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID                  uuid.UUID            `db:"id" json:"id"`
	WebauthnCredentials []WebauthnCredential `has_many:"webauthn_credentials" json:"webauthn_credentials,omitempty"`
	Emails              Emails               `has_many:"emails" json:"-"`
	Username            string               `db:"username" json:"username,omitempty"`
	CreatedAt           time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time            `db:"updated_at" json:"updated_at"`
}

func NewUser() User {
	id, _ := uuid.NewV4()
	return User{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (user *User) GetEmailById(emailId uuid.UUID) *Email {
	for _, email := range user.Emails {
		if email.ID.String() == emailId.String() {
			return &email
		}
	}
	return nil
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (user *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: user.ID},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: user.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: user.CreatedAt},
	), nil
}
