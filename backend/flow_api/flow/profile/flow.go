package profile

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateProfileInit                           flowpilot.StateName = "profile_init"
	StateProfileWebauthnCredentialVerification flowpilot.StateName = "webauthn_credential_verification"
	StateProfileAccountDeleted                 flowpilot.StateName = "account_deleted"
)

const (
	ActionAccountDelete                     flowpilot.ActionName = "account_delete"
	ActionEmailCreate                       flowpilot.ActionName = "email_create"
	ActionEmailDelete                       flowpilot.ActionName = "email_delete"
	ActionEmailVerify                       flowpilot.ActionName = "email_verify"
	ActionEmailSetPrimary                   flowpilot.ActionName = "email_set_primary"
	ActionPasswordSet                       flowpilot.ActionName = "password_set"
	ActionPasswordDelete                    flowpilot.ActionName = "password_delete"
	ActionUsernameSet                       flowpilot.ActionName = "username_set"
	ActionWebauthnCredentialCreate          flowpilot.ActionName = "webauthn_credential_create"
	ActionWebauthnCredentialRename          flowpilot.ActionName = "webauthn_credential_rename"
	ActionWebauthnCredentialDelete          flowpilot.ActionName = "webauthn_credential_delete"
	ActionWebauthnVerifyAttestationResponse flowpilot.ActionName = "webauthn_verify_attestation_response"
)

var Flow = flowpilot.NewFlow("/profile").
	InitialState(capabilities.StatePreflight, StateProfileInit).
	BeforeState(StateProfileInit, GetProfileData{}).
	State(StateProfileInit,
		AccountDelete{},
		EmailCreate{},
		EmailDelete{},
		EmailSetPrimary{},
		EmailVerify{},
		PasswordSet{},
		PasswordDelete{},
		UsernameSet{},
		WebauthnCredentialRename{},
		WebauthnCredentialCreate{},
		WebauthnCredentialDelete{},
	).
	State(StateProfileWebauthnCredentialVerification, WebauthnVerifyAttestationResponse{}).
	AfterState(StateProfileWebauthnCredentialVerification, WebauthnCredentialSave{}).
	AfterState(passcode.StatePasscodeConfirmation, shared.EmailPersistVerifiedStatus{}).
	State(StateProfileAccountDeleted).
	ErrorState(shared.StateError).
	SubFlows(capabilities.SubFlow, passcode.SubFlow).
	TTL(10 * time.Minute).
	Debug(true).
	MustBuild()
