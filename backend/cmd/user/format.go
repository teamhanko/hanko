package user

import (
	"errors"
	"fmt"
	"time"
)

type ImportEmail struct {
	Address    string `json:"address"`
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
}

type Emails []ImportEmail

type ImportEntry struct {
	UserID    string     `json:"user_id"`
	Emails    Emails     `json:"emails"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func validate(entries []ImportEntry) error {
	for i, e := range entries {
		if len(e.Emails) == 0 {
			return errors.New(fmt.Sprintf("Entry %v with id %v has got no Emails.", i, e.UserID))
		}
		primaryMails := 0
		for _, email := range e.Emails {
			if email.IsPrimary {
				primaryMails++
			}
		}
		if primaryMails != 1 {
			return errors.New(fmt.Sprintf("Need exactly one primary email, got %v", primaryMails))
		}
	}
	return nil
}
