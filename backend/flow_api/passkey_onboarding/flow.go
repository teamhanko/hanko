package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateOnboardingCreatePasskey            flowpilot.StateName = "onboarding_create_passkey"
	StateOnboardingVerifyPasskeyAttestation flowpilot.StateName = "onboarding_verify_passkey_attestation"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StateOnboardingCreatePasskey, GetWACreationOptions{}, Skip{}).
	State(StateOnboardingVerifyPasskeyAttestation, SendWAAttestationResponse{}).
	MustBuild()
