package flows

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func New2FACreationSubFlow() flowpilot.SubFlow {
	// TODO:
	return flowpilot.NewSubFlow().
		State(common.StateCreate2FASecurityKey).
		State(common.StateVerify2FASecurityKeyAssertion).
		State(common.StateCreate2FATOTP).
		State(common.StateGenerateRecoveryCodes).
		State(common.StateShowRecoveryCodes).
		FixedStates(common.StateCreate2FASecurityKey).
		MustBuild()
}
