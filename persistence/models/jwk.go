package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"time"
)

type Jwk struct {
	ID        int       `db:"id"`
	KeyData   string    `db:"key_data"`
	CreatedAt time.Time `db:"created_at"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (jwk *Jwk) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		//&validators.IntIsPresent{Name: "ID", Field: jwk.ID},
		&validators.StringIsPresent{Name: "KeyData", Field: jwk.KeyData},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: jwk.CreatedAt},
	), nil
}
