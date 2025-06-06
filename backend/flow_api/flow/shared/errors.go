package shared

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
	"net/http"
)

var (
	ErrorNotFound                         = flowpilot.NewFlowError("not_found", "The requested resource was not found.", http.StatusNotFound)
	ErrorPasscodeMaxAttemptsReached       = flowpilot.NewFlowError("passcode_max_attempts_reached", "The passcode was entered wrong too many times.", http.StatusUnauthorized)
	ErrorPasskeyInvalid                   = flowpilot.NewFlowError("passkey_invalid", "The passkey is invalid.", http.StatusUnauthorized)
	ErrorRateLimitExceeded                = flowpilot.NewFlowError("rate_limit_exceeded", "The rate limit has been exceeded.", http.StatusTooManyRequests)
	ErrorUnauthorized                     = flowpilot.NewFlowError("unauthorized", "The session is invalid.", http.StatusUnauthorized)
	ErrorWebauthnCredentialInvalidMFAOnly = flowpilot.NewFlowError("webauthn_credential_invalid_mfa_only", "This credential can be used as a second factor security key only.", http.StatusUnauthorized)
	ErrorPlatformAuthenticatorRequired    = flowpilot.NewFlowError("platform_authenticator_required", "The device or browser does not support the required platform authenticators.", http.StatusUnauthorized)
)

var (
	ErrorEmailAlreadyExists    = flowpilot.NewInputError("email_already_exists", "The email address already exists.")
	ErrorUsernameAlreadyExists = flowpilot.NewInputError("username_already_exists", "The username already exists.")
	ErrorUnknownUsername       = flowpilot.NewInputError("unknown_username_error", "The username is unknown.")
	ErrorUnknownEmail          = flowpilot.NewInputError("unknown_email_error", "The email address is unknown.")
	ErrorInvalidUsername       = flowpilot.NewInputError("invalid_username_error", "The username is invalid.")
	ErrorPasscodeInvalid       = flowpilot.NewInputError("passcode_invalid", "The passcode is invalid.")
	ErrorInvalidMetadata       = flowpilot.NewInputError("invalid_metadata_error", "The metadata is invalid.")
)
