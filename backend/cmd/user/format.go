package user

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

type ImportEmail struct {
	Address    string `json:"address" yaml:"address"`
	IsPrimary  bool   `json:"is_primary" yaml:"is_primary"`
	IsVerified bool   `json:"is_verified" yaml:"is_verified"`
}

type Emails []ImportEmail

type ImportEntry struct {
	UserID    string     `json:"user_id" yaml:"user_id"`
	Emails    Emails     `json:"emails" yaml:"emails"`
	CreatedAt *time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" yaml:"updated_at"`
}

func validateEntries(entries []ImportEntry) error {
	for i, e := range entries {
		if err := e.validate(); err != nil {
			return errors.Join(errors.New(fmt.Sprintf("Error with entry %v", i)), err)
		}
	}
	return nil
}

func (entry *ImportEntry) validate() error {
	if len(entry.Emails) == 0 {
		return errors.New(fmt.Sprintf("Entry with id: %v has got no Emails.", entry.UserID))
	}
	primaryMails := 0
	for _, email := range entry.Emails {
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
