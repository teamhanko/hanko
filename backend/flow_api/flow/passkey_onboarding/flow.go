package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
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

var SubFlow = flowpilot.NewSubFlow("passkey_onboarding").
	State(StateOnboardingCreatePasskey, WebauthnGenerateCreationOptions{}, Skip{}, Back{}).
	State(StateOnboardingVerifyPasskeyAttestation, WebauthnVerifyAttestationResponse{}, shared.Back{}).
	MustBuild()
