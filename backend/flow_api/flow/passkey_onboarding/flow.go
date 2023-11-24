package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateOnboardingCreatePasskey            flowpilot.StateName = "onboarding_create_passkey"
	StateOnboardingVerifyPasskeyAttestation flowpilot.StateName = "onboarding_verify_passkey_attestation"
)

const (
	ActionWebauthnGenerateCreationOptions   flowpilot.ActionName = "webauthn_generate_creation_options"
	ActionWebauthnVerifyAttestationResponse flowpilot.ActionName = "webauthn_verify_attestation_response"
	ActionSkip                              flowpilot.ActionName = "skip"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StateOnboardingCreatePasskey, WebauthnGenerateCreationOptions{}, Skip{}).
	State(StateOnboardingVerifyPasskeyAttestation, WebauthnVerifyAttestationResponse{}).
	MustBuild()
