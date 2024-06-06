package dto

import (
	"time"
)

type PasslinkFinishRequest struct {
	ID    string `json:"id" validate:"required,uuid4"`
	Token string `json:"token" validate:"required"`
}

type PasslinkInitRequest struct {
	UserID       string  `json:"user_id" validate:"required,uuid4"`
	EmailID      *string `json:"email_id"`
	RedirectPath string  `json:"redirect_path" validate:"required"`
}

type PasslinkReturn struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    string    `json:"user_id"`
}
