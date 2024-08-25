package admin

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type Passlink struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	EmailID    uuid.UUID `json:"email_id"`
	Email      *Email    `json:"email,omitempty"`
	TTL        int       `json:"ttl"` // in seconds
	LoginCount int       `json:"login_count"`
	Reusable   bool      `json:"reusable"` // by default a passlink can only used once, if reusable is set true, it can be used to authenticate the user multiple times by clicking the same link (e.g. in a newsletter)
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// FromPasslinkModel Converts the DB model to a DTO object
func FromPasslinkModel(model models.Passlink) Passlink {
	return Passlink{
		ID:         model.ID,
		UserID:     model.UserID,
		EmailID:    model.EmailID,
		Email:      FromEmailModel(&model.Email),
		TTL:        model.TTL,
		LoginCount: model.LoginCount,
		Reusable:   model.Reusable,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

type CreatePasslink struct {
	ID       *uuid.UUID `json:"id,omitempty"`
	UserID   uuid.UUID  `json:"user_id"`
	EmailID  uuid.UUID  `json:"email_id"`
	TTL      int        `json:"ttl"`      // in seconds
	Reusable bool       `json:"reusable"` // by default a passlink can only used once, if reusable is set true, it can be used to authenticate the user multiple times by clicking the same link (e.g. in a newsletter)
}
