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
	UserAgentRaw *string    `json:"user_agent_raw,omitempty"`
	UserAgent    *string    `json:"user_agent,omitempty"`
	IpAddress    *string    `json:"ip_address,omitempty"`
	Current      bool       `json:"current"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	LastUsed     time.Time  `json:"last_used"`
}

func FromSessionModel(model models.Session, current bool) SessionData {
	sessionData := SessionData{
		ID:        model.ID,
		Current:   current,
		CreatedAt: model.CreatedAt,
		ExpiresAt: model.ExpiresAt,
		LastUsed:  model.LastUsed,
	}

	if model.UserAgent.Valid {
		raw := model.UserAgent.String
		sessionData.UserAgentRaw = &raw
		ua := useragent.Parse(model.UserAgent.String)
		parsed := fmt.Sprintf("%s (%s)", ua.OS, ua.Name)
		sessionData.UserAgent = &parsed
	}

	if model.IpAddress.Valid {
		s := model.IpAddress.String
		sessionData.IpAddress = &s
	}

	return sessionData
}

type ValidateSessionResponse struct {
	IsValid        bool       `json:"is_valid"`
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
}

type ValidateSessionRequest struct {
	SessionToken string `json:"session_token" validate:"required"`
}
