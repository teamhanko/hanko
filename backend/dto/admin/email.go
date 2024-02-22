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

type CreateEmail struct {
	Address    string `json:"address" validate:"required,email"`
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
}

type EmailRequests interface {
	ListEmailRequestDto | CreateEmailRequestDto | GetEmailRequestDto
}

type ListEmailRequestDto struct {
	UserId string `param:"user_id" validate:"required,uuid4"`
}

type CreateEmailRequestDto struct {
	ListEmailRequestDto
	CreateEmail
}

type GetEmailRequestDto struct {
	ListEmailRequestDto
	EmailId string `param:"email_id" validate:"required,uuid4"`
}
