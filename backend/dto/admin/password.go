package admin

import (
	"github.com/gofrs/uuid"
	"time"
)

type PasswordCredential struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetPasswordCredentialRequestDto struct {
	UserID string `param:"user_id" validate:"required,uuid"`
}

type CreateOrUpdatePasswordCredentialRequestDto struct {
	GetPasswordCredentialRequestDto
	Password string `json:"password" validate:"required"`
}
