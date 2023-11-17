package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/actions"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/states"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

var SubFlow = flowpilot.NewSubFlow().
	State(states.StateOnboardingCreatePasskey, actions.GetWACreationOptions{}, sharedActions.Skip{}).
	State(states.StateOnboardingVerifyPasskeyAttestation, actions.SendWAAttestationResponse{}).
	MustBuild()
