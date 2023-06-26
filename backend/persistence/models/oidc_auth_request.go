package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"strings"
	"time"
)

type AuthRequest struct {
	ID    uuid.UUID  `db:"id" json:"id"`
	Codes []AuthCode `has_many:"codes" json:"codes,omitempty"`

	ClientID      string    `db:"client_id" json:"client_id"`
	CallbackURI   string    `db:"callback_uri" json:"callback_uri"`
	TransferState string    `db:"transfer_state" json:"transfer_state"`
	Prompt        string    `db:"prompt" json:"prompt"`
	UILocales     string    `db:"ui_locales" json:"ui_locales"`
	LoginHint     string    `db:"login_hint" json:"login_hint"`
	UserID        string    `db:"user_id" json:"user_id"`
	Scopes        string    `db:"scopes" json:"scopes"`
	ResponseType  string    `db:"response_type" json:"response_type"`
	Nonce         string    `db:"nonce" json:"nonce"`
	CodeChallenge string    `db:"code_challenge" json:"code_challenge"`
	MaxAuthAge    int64     `db:"max_auth_age" json:"max_auth_age"`
	Done          bool      `db:"done" json:"done"`
	AuthTime      time.Time `db:"auth_time" json:"auth_time"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func (t *AuthRequest) GetPrompt() []string {
	return strings.Split(t.Prompt, ",")
}

func (t *AuthRequest) GetUILocales() []string {
	return strings.Split(t.UILocales, ",")
}

func (t *AuthRequest) GetScopes() []string {
	return strings.Split(t.Scopes, ",")
}

func (t *AuthRequest) GetMaxAuthAge() time.Duration {
	return time.Duration(t.MaxAuthAge) * time.Second
}

func (t *AuthRequest) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: t.ID},
	), nil
}

type AuthCode struct {
	ID            string       `db:"id" json:"id"`
	AuthRequest   *AuthRequest `belongs_to:"auth_request" json:"auth_request,omitempty"`
	AuthRequestID uuid.UUID    `db:"auth_request_id" json:"auth_request_id"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
}
