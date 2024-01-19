package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"golang.org/x/exp/slices"
	"time"
)

// Email is used by pop to map your users database table to your go code.
type Email struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	UserID       *uuid.UUID    `db:"user_id" json:"user_id,omitempty"` // TODO: should not be a pointer anymore
	Address      string        `db:"address" json:"address"`
	Verified     bool          `db:"verified" json:"verified"`
	PrimaryEmail *PrimaryEmail `has_one:"primary_emails" json:"primary_emails,omitempty"`
	User         *User         `belongs_to:"user" json:"user,omitempty"`
	Identity     *Identity     `has_one:"identities" json:"identity,omitempty"`
	CreatedAt    time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `db:"updated_at" json:"updated_at"`
}

type Emails []Email

func NewEmail(userId *uuid.UUID, address string) *Email {
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
	if email.PrimaryEmail != nil && !email.PrimaryEmail.ID.IsNil() {
		return true
	}
	return false
}

func (emails Emails) GetVerified() Emails {
	var list Emails
	for _, email := range emails {
		if email.Verified {
			list = append(list, email)
		}
	}
	return list
}

func (emails Emails) HasUnverified() bool {
	return slices.ContainsFunc(emails, func(e Email) bool {
		return !e.Verified
	})
}

func (emails Emails) GetPrimary() *Email {
	for _, email := range emails {
		if email.IsPrimary() {
			return &email
		}
	}
	return nil
}

func (emails Emails) SetPrimary(primary *PrimaryEmail) {
	for i := range emails {
		if emails[i].ID.String() == primary.EmailID.String() {
			emails[i].PrimaryEmail = primary
			return
		}
	}
	return
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (email *Email) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: email.ID},
		&validators.EmailLike{Name: "Address", Field: email.Address},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: email.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: email.CreatedAt},
	), nil
}
