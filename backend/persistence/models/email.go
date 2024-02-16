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
	UserID       *uuid.UUID    `db:"user_id" json:"user_id,omitempty"`
	Address      string        `db:"address" json:"address"`
	Verified     bool          `db:"verified" json:"verified"`
	PrimaryEmail *PrimaryEmail `has_one:"primary_emails" json:"primary_emails,omitempty"`
	User         *User         `belongs_to:"user" json:"user,omitempty"`
	Identities   Identities    `has_many:"identities" json:"identity,omitempty"`
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

func (emails Emails) GetEmailByAddress(address string) *Email {
	for _, email := range emails {
		if email.Address == address {
			return &email
		}
	}
	return nil
}

func (emails Emails) GetEmailById(emailId uuid.UUID) *Email {
	for _, email := range emails {
		if email.ID.String() == emailId.String() {
			return &email
		}
	}
	return nil
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
