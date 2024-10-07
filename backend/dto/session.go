package dto

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type SessionData struct {
	ID        uuid.UUID `json:"id"`
	UserAgent string    `json:"user_agent"`
	IpAddress string    `json:"ip_address"`
	Current   bool      `json:"current"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func FromSessionModel(model models.Session, current bool) SessionData {
	return SessionData{
		ID:        model.ID,
		UserAgent: model.UserAgent,
		IpAddress: model.IpAddress,
		Current:   current,
		CreatedAt: model.CreatedAt,
		ExpiresAt: model.ExpiresAt,
	}
}
