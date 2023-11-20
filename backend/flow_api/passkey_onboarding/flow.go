package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateOnboardingCreatePasskey            flowpilot.StateName = "onboarding_create_passkey"
	StateOnboardingVerifyPasskeyAttestation flowpilot.StateName = "onboarding_verify_passkey_attestation"
)

const (
	ActionGetWACreationOptions      flowpilot.ActionName = "get_wa_creation_options"
	ActionSendWAAttestationResponse flowpilot.ActionName = "send_wa_attestation_response"
	ActionSkip                      flowpilot.ActionName = "skip"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StateOnboardingCreatePasskey, GetWACreationOptions{}, Skip{}).
	State(StateOnboardingVerifyPasskeyAttestation, SendWAAttestationResponse{}).
	MustBuild()
