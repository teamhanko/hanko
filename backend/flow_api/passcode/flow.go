package passcode

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared/hooks"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePasscodeConfirmation flowpilot.StateName = "passcode_confirmation"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePasscodeConfirmation, SubmitPasscode{}).
	BeforeState(StatePasscodeConfirmation, hooks.SendPasscode{}).
	MustBuild()
