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
