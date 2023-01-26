package admin

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type Email struct {
	ID         uuid.UUID `json:"id"`
	Address    string    `json:"address"`
	IsVerified bool      `json:"is_verified"`
	IsPrimary  bool      `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// FromEmailModel Converts the DB model to a DTO object
func FromEmailModel(email *models.Email) *Email {
	return &Email{
		ID:         email.ID,
		Address:    email.Address,
		IsVerified: email.Verified,
		IsPrimary:  email.IsPrimary(),
		CreatedAt:  email.CreatedAt,
		UpdatedAt:  email.UpdatedAt,
	}
}
