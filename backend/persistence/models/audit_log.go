package models

import (
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
	ActorEmail        *string      `db:"actor_email" json:"actor_email,omitempty"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
}

type AuditLogType string

var (
	AuditLogUserCreated   AuditLogType = "user_created"
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
	AuditLogWebAuthnCredentialUpdated            AuditLogType = "webauthn_credential_updated"
	AuditLogWebAuthnCredentialDeleted            AuditLogType = "webauthn_credential_deleted"

	AuditLogEmailCreated        AuditLogType = "email_created"
	AuditLogEmailDeleted        AuditLogType = "email_deleted"
	AuditLogEmailVerified       AuditLogType = "email_verified"
	AuditLogPrimaryEmailChanged AuditLogType = "primary_email_changed"
)
