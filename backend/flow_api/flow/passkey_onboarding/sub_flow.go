package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateIntroduction flowpilot.StateName = "passkey_onboarding_introduction"
	StateRegistration flowpilot.StateName = "passkey_onboarding_registration"
)

const (
	ActionWebauthnGenerateCreationOptions   flowpilot.ActionName = "webauthn_generate_creation_options"
	ActionWebauthnVerifyAttestationResponse flowpilot.ActionName = "webauthn_verify_attestation_response"
	ActionSkip                              flowpilot.ActionName = "skip"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StateIntroduction, WebauthnGenerateCreationOptions{}, Skip{}).
	State(StateRegistration, WebauthnVerifyAttestationResponse{}).
	MustBuild()
