package user

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

// ImportEmail The import format for a user's email
type ImportEmail struct {
	// Address Valid email address
	Address string `json:"address" yaml:"address"`
	// IsPrimary indicates if this is the primary email of the users. In the Emails array there has to be exactly one Primary EMail.
	IsPrimary bool `json:"is_primary" yaml:"is_primary"`
	// IsVerified indicates if the email address was previously verified.
	IsVerified bool `json:"is_verified" yaml:"is_verified"`
}

// Emails Array of Email Addresses
type Emails []ImportEmail

// ImportEntry represents a user to be imported to the hanko database
type ImportEntry struct {
	// UserID optional uuid.v4. If not provided a new one will be generated for the user
	UserID string `json:"user_id" yaml:"user_id"`
	// Emails List of emails
	Emails Emails `json:"emails" yaml:"emails"`
	// CreatedAt optional timestamp of the users' creation. Will be set to the import date if not provided.
	CreatedAt *time.Time `json:"created_at" yaml:"created_at"`
	// UpdatedAt optional timestamp of the last update to the user. Will be set to the import date if not provided.
	UpdatedAt *time.Time `json:"updated_at" yaml:"updated_at"`
}

// ImportList a list of ImportEntries
type ImportList []ImportEntry

func (entry *ImportEntry) validate() error {
	if len(entry.Emails) == 0 {
		return errors.New(fmt.Sprintf("Entry with id: %v has got no Emails.", entry.UserID))
	}
	primaryMails := 0
	for _, email := range entry.Emails {
		//TODO: Validate email
		if email.IsPrimary {
			primaryMails++
		}
	}

	if primaryMails != 1 {
		return errors.New(fmt.Sprintf("Need exactly one primary email, got %v", primaryMails))
	}
	if entry.UserID != "" {
		_, err := uuid.FromString(entry.UserID)
		if err != nil {
			return errors.New(fmt.Sprintf("Provided uuid is not valid: %v", entry.UserID))
		}
	}
	return nil
}
