package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// Email is used by pop to map your users database table to your go code.
type Email struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	UserID       uuid.UUID     `db:"user_id" json:"user_id"`
	Address      string        `db:"address" json:"address"`
	Verified     bool          `db:"verified" json:"verified"`
	PrimaryEmail *PrimaryEmail `has_one:"primary_emails" json:"primary_emails,omitempty"`
	User         *User         `belongs_to:"tree" json:"user,omitempty"`
	CreatedAt    time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `db:"updated_at" json:"updated_at"`
}

func NewEmail(userId uuid.UUID, address string) *Email {
	id, _ := uuid.NewV4()
	return &Email{
		ID:           id,
		Address:      address,
		UserID:       userId,
		Verified:     false,
		PrimaryEmail: nil,
		User:         nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (email *Email) IsPrimary() bool {
	if email.PrimaryEmail != nil {
		return true
	}
	return false
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (email *Email) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: email.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: email.UserID},
		&validators.EmailLike{Name: "Address", Field: email.Address},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: email.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: email.CreatedAt},
	), nil
}
