package models

import (
	"encoding/base64"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/nulls"
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
	ExpiresAt          nulls.Time                             `db:"expires_at"`
}

func (sd *WebauthnSessionData) decodeAllowedCredentials() [][]byte {
	var allowedCredentials [][]byte

	for _, credential := range sd.AllowedCredentials {
		credentialId, err := base64.RawURLEncoding.DecodeString(credential.CredentialId)
		if err != nil {
			continue
		}

		allowedCredentials = append(allowedCredentials, credentialId)
	}

	return allowedCredentials
}

func NewWebauthnSessionDataFrom(sessionData *webauthn.SessionData, operation Operation) (*WebauthnSessionData, error) {
	now := time.Now().UTC()

	sessionDataID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate a new uuid for session data: %w", err)
	}

	userID, _ := uuid.FromBytes(sessionData.UserID)

	allowedCredentials := make([]WebauthnSessionDataAllowedCredential, len(sessionData.AllowedCredentialIDs))

	for index, credentialID := range sessionData.AllowedCredentialIDs {
		allowedCredentialID, err := uuid.NewV4()
		if err != nil {
			return nil, fmt.Errorf("failed to generate a uuid for the allowed credential: %w", err)
		}

		allowedCredential := WebauthnSessionDataAllowedCredential{
			ID:                    allowedCredentialID,
			CredentialId:          base64.RawURLEncoding.EncodeToString(credentialID),
			WebauthnSessionDataID: sessionDataID,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		allowedCredentials[index] = allowedCredential
	}

	sessionDataModel := &WebauthnSessionData{
		ID:                 sessionDataID,
		Challenge:          sessionData.Challenge,
		UserId:             userID,
		UserVerification:   string(sessionData.UserVerification),
		CreatedAt:          now,
		UpdatedAt:          now,
		Operation:          operation,
		AllowedCredentials: allowedCredentials,
		ExpiresAt:          nulls.NewTime(sessionData.Expires),
	}

	return sessionDataModel, nil
}

func (sd *WebauthnSessionData) ToSessionData() *webauthn.SessionData {
	allowedCredentials := sd.decodeAllowedCredentials()

	// TODO: do we need the following lines and is the user optional?
	var userId []byte = nil

	if !sd.UserId.IsNil() {
		userId = sd.UserId.Bytes()
	}

	sessionData := &webauthn.SessionData{
		Challenge:            sd.Challenge,
		UserID:               userId,
		AllowedCredentialIDs: allowedCredentials,
		UserVerification:     protocol.UserVerificationRequirement(sd.UserVerification),
		Expires:              sd.ExpiresAt.Time,
	}

	return sessionData
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
