package dto

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/mileusna/useragent"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type SessionData struct {
	ID           uuid.UUID  `json:"id"`
	UserAgentRaw string     `json:"user_agent_raw"`
	UserAgent    string     `json:"user_agent"`
	IpAddress    string     `json:"ip_address"`
	Current      bool       `json:"current"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	LastUsed     time.Time  `json:"last_used"`
}

func FromSessionModel(model models.Session, current bool) SessionData {
	ua := useragent.Parse(model.UserAgent)
	return SessionData{
		ID:           model.ID,
		UserAgentRaw: model.UserAgent,
		UserAgent:    fmt.Sprintf("%s (%s)", ua.OS, ua.Name),
		IpAddress:    model.IpAddress,
		Current:      current,
		CreatedAt:    model.CreatedAt,
		ExpiresAt:    model.ExpiresAt,
		LastUsed:     model.LastUsed,
	}
}

type ValidateSessionResponse struct {
	IsValid        bool       `json:"is_valid"`
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
}

type ValidateSessionRequest struct {
	SessionToken string `json:"session_token" validate:"required"`
}
