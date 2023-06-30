package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"strings"
	"time"
)

type AccessToken struct {
	ID             uuid.UUID     `db:"id" json:"id"`
	RefreshToken   *RefreshToken `belongs_to:"refresh_tokens" json:"refresh_token,omitempty"`
	RefreshTokenID *uuid.UUID    `db:"refresh_token_id" json:"refresh_token_id"`

	ClientID  string    `db:"client_id" json:"client_id"`
	Subject   string    `db:"subject" json:"subject"`
	Audience  string    `db:"audience" json:"audience"`
	Scopes    string    `db:"scopes" json:"scopes"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (t *AccessToken) GetAudience() []string {
	return strings.Split(t.ClientID, ",")
}

func (t *AccessToken) GetScopes() []string {
	return strings.Split(t.Scopes, ",")
}

func (t *AccessToken) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: t.ID},
		&validators.StringIsPresent{Name: "ClientID", Field: t.ClientID},
		&validators.StringIsPresent{Name: "Subject", Field: t.Subject},
		&validators.TimeIsPresent{Name: "Expires At", Field: t.ExpiresAt},
	), nil
}

type RefreshToken struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	AccessTokens []AccessToken `has_many:"access_tokens" json:"access_tokens,omitempty"`

	ClientID  string    `db:"client_id" json:"client_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Audience  string    `db:"audience" json:"audience"`
	AMR       string    `db:"amr" json:"amr"`
	Scopes    string    `db:"scopes" json:"scopes"`
	AuthTime  time.Time `db:"auth_time" json:"auth_time"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (t *RefreshToken) GetAudience() []string {
	return strings.Split(t.ClientID, ",")
}

func (t *RefreshToken) GetScopes() []string {
	return strings.Split(t.Scopes, ",")
}

func (t *RefreshToken) GetAMR() []string {
	return strings.Split(t.AMR, ",")
}

func (t *RefreshToken) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: t.ID},
		&validators.StringIsPresent{Name: "ClientID", Field: t.ClientID},
		&validators.TimeIsPresent{Name: "AuthTime", Field: t.AuthTime},
		&validators.TimeIsPresent{Name: "ExpiresAt", Field: t.ExpiresAt},
	), nil
}
