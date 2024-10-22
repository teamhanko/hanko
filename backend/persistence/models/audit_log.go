package models

import (
	"fmt"
	"github.com/gobuffalo/pop/v6/slices"
	"github.com/gofrs/uuid"
	"time"
)

type AuditLog struct {
	ID                uuid.UUID    `db:"id" json:"id"`
	Type              AuditLogType `db:"type" json:"type"`
	Error             *string      `db:"error" json:"error,omitempty"`
	MetaHttpRequestId string       `db:"meta_http_request_id" json:"meta_http_request_id"`
	MetaSourceIp      string       `db:"meta_source_ip" json:"meta_source_ip"`
	MetaUserAgent     string       `db:"meta_user_agent" json:"meta_user_agent"`
	ActorUserId       *uuid.UUID   `db:"actor_user_id" json:"actor_user_id,omitempty"`
	ActorEmail        *string      `db:"actor_email" json:"actor_email,omitempty" mask:"email"`
	Details           slices.Map   `db:"details" json:"details"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
}

type Details map[string]interface{}

type RequestMeta struct {
	HttpRequestId string
	SourceIp      string
	UserAgent     string
}

func NewAuditLog(auditLogType AuditLogType, requestMeta RequestMeta, details Details, user *User, logError error) (AuditLog, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return AuditLog{}, fmt.Errorf("failed to create id: %w", err)
	}

	auditLog := AuditLog{
		ID:                id,
		Type:              auditLogType,
		Error:             nil,
		MetaHttpRequestId: requestMeta.HttpRequestId,
		MetaUserAgent:     requestMeta.UserAgent,
		MetaSourceIp:      requestMeta.SourceIp,
		ActorUserId:       nil,
		ActorEmail:        nil,
	}

	if len(details) > 0 {
		auditLog.Details = slices.Map(details)
	}

	if user != nil {
		auditLog.ActorUserId = &user.ID

		if e := user.Emails.GetPrimary(); e != nil {
			auditLog.ActorEmail = &e.Address
		}
	}

	if logError != nil {
		// check if error is not nil, because else the string (formatted with fmt.Sprintf) would not be empty but look like this: `%!s(<nil>)`
		tmp := fmt.Sprintf("%s", logError)
		auditLog.Error = &tmp
	}
	return auditLog, nil
}

type AuditLogType string

var (
	AuditLogUserLoggedOut AuditLogType = "user_logged_out"

	AuditLogPasswordSetSucceeded AuditLogType = "password_set_succeeded"
	AuditLogPasswordSetFailed    AuditLogType = "password_set_failed"

	AuditLogPasswordLoginSucceeded AuditLogType = "password_login_succeeded"
	AuditLogPasswordLoginFailed    AuditLogType = "password_login_failed"

	AuditLogPasscodeLoginInitSucceeded  AuditLogType = "passcode_login_init_succeeded"
	AuditLogPasscodeLoginInitFailed     AuditLogType = "passcode_login_init_failed"
	AuditLogPasscodeLoginFinalSucceeded AuditLogType = "passcode_login_final_succeeded"
	AuditLogPasscodeLoginFinalFailed    AuditLogType = "passcode_login_final_failed"

	AuditLogWebAuthnRegistrationInitSucceeded  AuditLogType = "webauthn_registration_init_succeeded"
	AuditLogWebAuthnRegistrationInitFailed     AuditLogType = "webauthn_registration_init_failed"
	AuditLogWebAuthnRegistrationFinalSucceeded AuditLogType = "webauthn_registration_final_succeeded"
	AuditLogWebAuthnRegistrationFinalFailed    AuditLogType = "webauthn_registration_final_failed"

	AuditLogWebAuthnAuthenticationInitSucceeded  AuditLogType = "webauthn_authentication_init_succeeded"
	AuditLogWebAuthnAuthenticationInitFailed     AuditLogType = "webauthn_authentication_init_failed"
	AuditLogWebAuthnAuthenticationFinalSucceeded AuditLogType = "webauthn_authentication_final_succeeded"
	AuditLogWebAuthnAuthenticationFinalFailed    AuditLogType = "webauthn_authentication_final_failed"

	AuditLogWebAuthnCredentialUpdated AuditLogType = "webauthn_credential_updated"
	AuditLogWebAuthnCredentialDeleted AuditLogType = "webauthn_credential_deleted"

	AuditLogThirdPartySignUpSucceeded    AuditLogType = "thirdparty_signup_succeeded"
	AuditLogThirdPartySignInSucceeded    AuditLogType = "thirdparty_signin_succeeded"
	AuditLogThirdPartyLinkingSucceeded   AuditLogType = "thirdparty_linking_succeeded"
	AuditLogThirdPartySignInSignUpFailed AuditLogType = "thirdparty_signin_signup_failed"

	AuditLogTokenExchangeSucceeded AuditLogType = "token_exchange_succeeded"
	AuditLogTokenExchangeFailed    AuditLogType = "token_exchange_failed"

	// Types used by old API and new/flow API
	AuditLogUserCreated         AuditLogType = "user_created"
	AuditLogEmailCreated        AuditLogType = "email_created"
	AuditLogEmailVerified       AuditLogType = "email_verified"
	AuditLogEmailDeleted        AuditLogType = "email_deleted"
	AuditLogPrimaryEmailChanged AuditLogType = "primary_email_changed"
	AuditLogUserDeleted         AuditLogType = "user_deleted"

	// New/flow API types
	AuditLogLoginSuccess       AuditLogType = "login_success"
	AuditLogLoginFailure       AuditLogType = "login_failure"
	AuditLogOTPCreated         AuditLogType = "otp_created"
	AuditLogPasskeyCreated     AuditLogType = "passkey_created"
	AuditLogPasskeyDeleted     AuditLogType = "passkey_deleted"
	AuditLogSecurityKeyCreated AuditLogType = "security_key_created"
	AuditLogUsernameChanged    AuditLogType = "username_changed"
	AuditLogUsernameDeleted    AuditLogType = "username_deleted"
	AuditLogPasswordChanged    AuditLogType = "password_changed"
	AuditLogPasswordDeleted    AuditLogType = "password_deleted"
)
