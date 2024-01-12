package models

import (
	"github.com/gobuffalo/validate/v3/validators"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Webhook is used by pop to map your webhooks database table to your go code.
type Webhook struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	Callback      string        `json:"callback" db:"callback"`
	Enabled       bool          `json:"enabled" db:"enabled"`
	Failures      int           `json:"failures" db:"failures"`
	ExpiresAt     time.Time     `json:"expires_at" db:"expires_at"`
	WebhookEvents WebhookEvents `json:"events" has_many:"webhook_events"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// Webhooks are not required by pop and may be deleted
type Webhooks []Webhook

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (w *Webhook) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: w.ID},
		&validators.StringIsPresent{Name: "Callback", Field: w.Callback},
		&validators.TimeIsPresent{Name: "ExpiresAt", Field: w.ExpiresAt},
		&validators.TimeAfterTime{
			FirstName:  "Expires At",
			FirstTime:  w.ExpiresAt,
			SecondName: "Now",
			SecondTime: time.Now(),
		},

		&validators.TimeIsPresent{Name: "UpdatedAt", Field: w.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: w.CreatedAt},
	), nil
}
