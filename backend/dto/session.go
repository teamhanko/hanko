package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/mileusna/useragent"
	"github.com/teamhanko/hanko/backend/persistence/models"
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

type Claims struct {
	Subject      uuid.UUID              `json:"subject"`
	IssuedAt     *time.Time             `json:"issued_at,omitempty"`
	Expiration   time.Time              `json:"expiration"`
	Audience     []string               `json:"audience,omitempty"`
	Issuer       *string                `json:"issuer,omitempty"`
	Email        *EmailJWT              `json:"email,omitempty"`
	Username     *string                `json:"username,omitempty"`
	SessionID    uuid.UUID              `json:"session_id"`
	CustomClaims map[string]interface{} `json:"-"`
}

// Custom MarshalJSON to flatten CustomClaims into the top level
func (c Claims) MarshalJSON() ([]byte, error) {
	// Create a map to hold the flattened structure
	flattened := make(map[string]interface{})

	// Marshal basic fields into the flattened map
	flattened["subject"] = c.Subject
	flattened["expiration"] = c.Expiration
	flattened["session_id"] = c.SessionID

	if c.IssuedAt != nil {
		flattened["issued_at"] = c.IssuedAt
	}
	if len(c.Audience) > 0 {
		flattened["audience"] = c.Audience
	}
	if c.Issuer != nil {
		flattened["issuer"] = c.Issuer
	}
	if c.Email != nil {
		flattened["email"] = c.Email
	}
	if c.Username != nil {
		flattened["username"] = c.Username
	}

	// Flatten CustomClaims into the top level
	for key, value := range c.CustomClaims {
		flattened[key] = value
	}

	return json.Marshal(flattened)
}

func GetClaimsFromToken(token jwt.Token) (*Claims, error) {
	claims := &Claims{
		CustomClaims: make(map[string]interface{}),
	}

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

	if username, valid := token.Get("username"); valid {
		if usernameStr, validStr := username.(string); validStr {
			claims.Username = &usernameStr
		}
	}

	claims.Expiration = token.Expiration()

	hankoClaims := map[string]bool{
		"email":      true,
		"username":   true,
		"session_id": true,
	}

	for key, value := range token.PrivateClaims() {
		if !hankoClaims[key] {
			claims.CustomClaims[key] = value
		}
	}

	return claims, nil
}

type ValidateSessionResponse struct {
	IsValid bool    `json:"is_valid"`
	Claims  *Claims `json:"claims,omitempty"`
	// deprecated
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
	// deprecated
	UserID *uuid.UUID `json:"user_id,omitempty"`
}

type ValidateSessionRequest struct {
	SessionToken string `json:"session_token" validate:"required"`
}
