package _fa_creation

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func New2FACreationSubFlow() flowpilot.SubFlow {
	// TODO:
	return flowpilot.NewSubFlow().
		State(shared.StateCreate2FASecurityKey).
		State(shared.StateVerify2FASecurityKeyAssertion).
		State(shared.StateCreate2FATOTP).
		State(shared.StateGenerateRecoveryCodes).
		State(shared.StateShowRecoveryCodes).
		MustBuild()
}
