package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Passlink is used by pop to map your passlink database table to your go code.
type Passlink struct {
	ID         uuid.UUID `db:"id"`
	UserId     uuid.UUID `db:"user_id"`
	EmailID    uuid.UUID `db:"email_id"`
	TTL        int       `db:"ttl"` // in seconds
	IP         string    `db:"ip"`
	Token      string    `db:"token"`
	LoginCount int       `db:"login_count"`
	Reusable   bool      `db:"reusable"` // by default a passlink can only used once, if reusable is set true, it can be used to authenticate the user multiple times by clicking the same link (e.g. in a newsletter)
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	Email      Email     `belongs_to:"email"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (passlink *Passlink) Validate(tx *pop.Connection) (*validate.Errors, error) {
	tests := []validate.Validator{
		&validators.UUIDIsPresent{Name: "ID", Field: passlink.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: passlink.UserId},
		&validators.StringLengthInRange{Name: "Token", Field: passlink.Token, Min: 16},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: passlink.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: passlink.UpdatedAt},
	}
	return validate.Validate(tests...), nil
}
