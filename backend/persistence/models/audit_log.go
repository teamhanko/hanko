package models

import (
	"github.com/gofrs/uuid"
	"time"
)

type AuditLog struct {
	ID                uuid.UUID    `db:"id" json:"id"`
	Type              AuditLogType `db:"type" json:"type"`
	Error             string       `db:"error" json:"error"`
	MetaHttpRequestId string       `db:"meta_http_request_id" json:"meta_http_request_id"`
	MetaSourceIp      string       `db:"meta_source_ip" json:"meta_source_ip"`
	MetaUserAgent     string       `db:"meta_user_agent" json:"meta_user_agent"`
	ActorUserId       *uuid.UUID   `db:"actor_user_id" json:"actor_user_id"`
	ActorEmail        string       `db:"actor_email" json:"actor_email"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
}

type AuditLogType string

var (
	AuditLogUserCreated AuditLogType = "user_created"

	AuditLogPasswordSetSucceeded AuditLogType = "password_set_succeeded"
	AuditLogPasswordSetFailed    AuditLogType = "password_set_failed"

	AuditLogPasswordLoginSucceeded AuditLogType = "password_login_succeeded"
	AuditLogPasswordLoginFailed    AuditLogType = "password_login_failed"

	AuditLogPasscodeLoginInitSucceeded AuditLogType = "passcode_login_init_succeeded"
	AuditLogPasscodeLoginInitFailed    AuditLogType = "passcode_login_init_failed"
	AuditLogPasscodeLoginSucceeded     AuditLogType = "passcode_login_succeeded"
	AuditLogPasscodeLoginFailed        AuditLogType = "passcode_login_failed"

	AuditLogWebAuthnRegistrationInitSucceeded AuditLogType = "webauthn_registration_init_succeeded"
	AuditLogWebAuthnRegistrationInitFailed    AuditLogType = "webauthn_registration_init_failed"
	AuditLogWebAuthnRegistrationSucceeded     AuditLogType = "webauthn_registration_succeeded"
	AuditLogWebAuthnRegistrationFailed        AuditLogType = "webauthn_registration_failed"

	AuditLogWebAuthnAuthenticationInitSucceeded AuditLogType = "webauthn_authentication_init_succeeded"
	AuditLogWebAuthnAuthenticationInitFailed    AuditLogType = "webauthn_authentication_init_failed"
	AuditLogWebAuthnAuthenticationSucceeded     AuditLogType = "webauthn_authentication_succeeded"
	AuditLogWebAuthnAuthenticationFailed        AuditLogType = "webauthn_authentication_failed"
)
