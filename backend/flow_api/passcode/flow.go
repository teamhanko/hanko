package passcode

import (
	"github.com/teamhanko/hanko/backend/flow_api/passcode/actions"
	"github.com/teamhanko/hanko/backend/flow_api/passcode/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared/hooks"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

var SubFlow = flowpilot.NewSubFlow().
	State(states.StatePasscodeConfirmation, actions.SubmitPasscode{}).
	BeforeState(states.StatePasscodeConfirmation, hooks.SendPasscode{}).
	MustBuild()
