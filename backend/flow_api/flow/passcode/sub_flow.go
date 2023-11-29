package passcode

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateConfirmation flowpilot.StateName = "passcode_confirmation"
)

const (
	ActionConfirmPasscode flowpilot.ActionName = "confirm_passcode"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StateConfirmation, ConfirmPasscode{}).
	BeforeState(StateConfirmation, SendPasscode{}).
	MustBuild()
