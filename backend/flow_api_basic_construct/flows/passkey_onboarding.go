package flows

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/actions"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewPasskeyOnboardingSubFlow() flowpilot.SubFlow {
	// TODO:
	return flowpilot.NewSubFlow().
		State(common.StateOnboardingCreatePasskey, actions.NewGetWACreationOptions(), actions.NewSkip()).
		State(common.StateOnboardingVerifyPasskeyAttestation, actions.NewSendWAAttestationResponse()).
		FixedStates(common.StateOnboardingCreatePasskey).
		MustBuild()
}
