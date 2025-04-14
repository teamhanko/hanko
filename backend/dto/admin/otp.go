package admin

import (
	"github.com/gofrs/uuid"
	"time"
)

type GetOTPRequestDto struct {
	UserID string `param:"user_id" validate:"required,uuid"`
}

type OTPDto struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
