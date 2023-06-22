package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type AuthRequest struct {
	ID    uuid.UUID  `db:"id" json:"id"`
	Codes []AuthCode `has_many:"codes" json:"codes,omitempty"`

	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
	ClientID      string        `db:"client_id" json:"client_id"`
	CallbackURI   string        `db:"callback_uri" json:"callback_uri"`
	TransferState string        `db:"transfer_state" json:"transfer_state"`
	Prompt        []string      `db:"prompt" json:"prompt"`
	UILocales     []string      `db:"ui_locales" json:"ui_locales"`
	LoginHint     string        `db:"login_hint" json:"login_hint"`
	MaxAuthAge    time.Duration `db:"max_auth_age" json:"max_auth_age"`
	UserID        string        `db:"user_id" json:"user_id"`
	Scopes        []string      `db:"scopes" json:"scopes"`
	ResponseType  string        `db:"response_type" json:"response_type"`
	Nonce         string        `db:"nonce" json:"nonce"`
	CodeChallenge string        `db:"code_challenge" json:"code_challenge"`
}

func (t *AuthRequest) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: t.ID},
	), nil
}

type AuthCode struct {
	ID          string       `db:"id" json:"id"`
	AuthRequest *AuthRequest `belongs_to:"auth_requests" json:"auth_request,omitempty"`
}
