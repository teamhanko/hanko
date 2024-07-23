package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type Username struct {
	ID        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromUsernameModel(u *models.Username) *Username {
	if u == nil {
		return nil
	}
	return &Username{
		ID:        u.ID,
		UserId:    u.UserId,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
