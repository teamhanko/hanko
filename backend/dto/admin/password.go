package admin

import (
	"time"

	"github.com/gofrs/uuid"
)

type PasswordCredential struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
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
