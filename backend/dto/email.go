package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailResponse struct {
	ID         uuid.UUID `json:"id"`
	Address    string    `json:"address"`
	IsVerified bool      `json:"is_verified"`
	IsPrimary  bool      `json:"is_primary"`
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
		ID:         email.ID,
		Address:    email.Address,
		IsVerified: email.Verified,
		IsPrimary:  email.IsPrimary(),
	}
}
