package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type EmailResponse struct {
	ID        uuid.UUID `json:"id"`
	Address   string    `json:"address"`
	Verified  bool      `json:"verified"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

type EmailCreateRequest struct {
	Address string `json:"address"`
}

type EmailUpdateRequest struct {
	IsPrimary *bool `json:"is_primary"`
}

// FromEmailModel Converts the DB model to a DTO object
func FromEmailModel(email *models.Email) *EmailResponse {
	return &EmailResponse{
		ID:        email.ID,
		Address:   email.Address,
		Verified:  email.Verified,
		IsPrimary: email.IsPrimary(),
		CreatedAt: time.Time{},
	}
}
