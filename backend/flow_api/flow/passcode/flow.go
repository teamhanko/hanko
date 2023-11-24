package passcode

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePasscodeConfirmation flowpilot.StateName = "passcode_confirmation"
)

const (
	ActionVerifyPasscode flowpilot.ActionName = "verify_passcode"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePasscodeConfirmation, VerifyPasscode{}).
	BeforeState(StatePasscodeConfirmation, SendPasscode{}).
	MustBuild()
