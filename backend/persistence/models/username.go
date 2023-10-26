package models

import (
	"github.com/gofrs/uuid"
	"time"
)

type Username struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at", json:"updated_at"`
}

func NewUsername(userId uuid.UUID, username string) *Username {
	id, _ := uuid.NewV4()
	return &Username{
		ID:       id,
		Username: username,
		UserID:   userId,
	}
}
