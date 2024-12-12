package dto

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
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

type Claims struct {
	Subject    uuid.UUID  `json:"subject"`
	IssuedAt   *time.Time `json:"issued_at,omitempty"`
	Expiration time.Time  `json:"expiration"`
	Audience   []string   `json:"audience,omitempty"`
	Issuer     *string    `json:"issuer,omitempty"`
	Email      *EmailJwt  `json:"email,omitempty"`
	SessionID  uuid.UUID  `json:"session_id"`
}

type ValidateSessionResponse struct {
	IsValid bool    `json:"is_valid"`
	Claims  *Claims `json:"claims,omitempty"`
	// deprecated
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
	// deprecated
	UserID *uuid.UUID `json:"user_id,omitempty"`
}

func GetClaimsFromToken(token jwt.Token) (*Claims, error) {
	claims := &Claims{}

	if subject := token.Subject(); len(subject) > 0 {
		s, err := uuid.FromString(subject)
		if err != nil {
			return nil, fmt.Errorf("'subject' is not a uuid: %w", err)
		}
		claims.Subject = s
	}

	if sessionID, valid := token.Get("session_id"); valid {
		s, err := uuid.FromString(sessionID.(string))
		if err != nil {
			return nil, fmt.Errorf("'session_id' is not a uuid: %w", err)
		}
		claims.SessionID = s
	}

	if issuedAt := token.IssuedAt(); !issuedAt.IsZero() {
		claims.IssuedAt = &issuedAt
	}

	if audience := token.Audience(); len(audience) > 0 {
		claims.Audience = audience
	}

	if issuer := token.Issuer(); len(issuer) > 0 {
		claims.Issuer = &issuer
	}

	if email, valid := token.Get("email"); valid {
		if data, ok := email.(map[string]interface{}); ok {
			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal 'email' claim: %w", err)
			}
			err = json.Unmarshal(jsonData, &claims.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal 'email' claim: %w", err)
			}
		}
	}

	claims.Expiration = token.Expiration()

	return claims, nil
}

type ValidateSessionRequest struct {
	SessionToken string `json:"session_token" validate:"required"`
}
