package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type AccessToken struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	RefreshToken *RefreshToken `belongs_to:"refresh_tokens" json:"refresh_token,omitempty"`

	ClientID   string    `db:"client_id" json:"client_id"`
	Subject    string    `db:"subject" json:"subject"`
	Audience   []string  `db:"audience" json:"audience"`
	Expiration time.Time `db:"expiration" json:"expiration"`
	Scopes     []string  `db:"scopes" json:"scopes"`
}

func (t *AccessToken) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: t.ID},
		&validators.StringIsPresent{Name: "ClientID", Field: t.ClientID},
		&validators.StringIsPresent{Name: "Subject", Field: t.Subject},
		&validators.TimeIsPresent{Name: "Expiration", Field: t.Expiration},
	), nil
}

type RefreshToken struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	AccessTokens []AccessToken `has_many:"access_tokens" json:"access_tokens,omitempty"`

	ClientID   string    `db:"client_id" json:"client_id"`
	Audience   []string  `db:"audience" json:"audience"`
	AuthTime   time.Time `db:"auth_time" json:"auth_time"`
	AMR        []string  `db:"amr" json:"amr"`
	Scopes     []string  `db:"scopes" json:"scopes"`
	UserID     string    `db:"user_id" json:"user_id"`
	Expiration time.Time `db:"expiration" json:"expiration"`
}

func (t *RefreshToken) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: t.ID},
		&validators.StringIsPresent{Name: "ClientID", Field: t.ClientID},
		&validators.TimeIsPresent{Name: "AuthTime", Field: t.AuthTime},
		&validators.TimeIsPresent{Name: "Expiration", Field: t.Expiration},
	), nil
}
