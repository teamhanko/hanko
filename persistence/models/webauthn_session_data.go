package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type Operation string

var (
	WebauthnOperationRegistration   Operation = "registration"
	WebauthnOperationAuthentication Operation = "authentication"
)

// WebauthnSessionData is used by pop to map your webauthn_session_data database table to your go code.
type WebauthnSessionData struct {
	ID                 uuid.UUID                              `db:"id"`
	Challenge          string                                 `db:"challenge"`
	UserId             uuid.UUID                              `db:"user_id"`
	UserVerification   string                                 `db:"user_verification"`
	CreatedAt          time.Time                              `db:"created_at"`
	UpdatedAt          time.Time                              `db:"updated_at"`
	Operation          Operation                              `db:"operation"`
	AllowedCredentials []WebauthnSessionDataAllowedCredential `has_many:"webauthn_session_data_allowed_credentials"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (sd *WebauthnSessionData) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: sd.ID},
		&validators.StringIsPresent{Name: "Challenge", Field: sd.Challenge},
		&validators.StringInclusion{Name: "Operation", Field: string(sd.Operation), List: []string{string(WebauthnOperationRegistration), string(WebauthnOperationAuthentication)}},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: sd.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: sd.CreatedAt},
	), nil
}
