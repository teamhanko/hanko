package models

import (
	"github.com/gobuffalo/validate/v3/validators"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// WebhookEvent is used by pop to map your webhook_events database table to your go code.
type WebhookEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Webhook   *Webhook  `json:"-" belongs_to:"webhook"`
	WebhookID uuid.UUID `json:"-" db:"webhook_id"`
	Event     string    `json:"event" db:"event"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

// WebhookEvents is not required by pop and may be deleted
type WebhookEvents []WebhookEvent

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (w *WebhookEvent) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: w.ID},
		&validators.StringIsPresent{Name: "Event", Field: w.Event},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: w.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: w.CreatedAt},
	), nil
}
