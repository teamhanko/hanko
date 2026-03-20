package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

type Tenant struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	Config    json.RawMessage `db:"config" json:"config"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

func (t *Tenant) BeforeCreate(tx *pop.Connection) error {
	if t.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		t.ID = id
	}
	return nil
}

func (t *Tenant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
