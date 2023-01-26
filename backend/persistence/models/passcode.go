package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

// Passcode is used by pop to map your passcodes database table to your go code.
type Passcode struct {
	ID        uuid.UUID `db:"id"`
	UserId    uuid.UUID `db:"user_id"`
	EmailID   uuid.UUID `db:"email_id"`
	Ttl       int       `db:"ttl"` // in seconds
	Code      string    `db:"code"`
	TryCount  int       `db:"try_count"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Email     Email     `belongs_to:"email"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (passcode *Passcode) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: passcode.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: passcode.UserId},
		&validators.StringLengthInRange{Name: "Code", Field: passcode.Code, Min: 6},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: passcode.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: passcode.UpdatedAt},
	), nil
}
