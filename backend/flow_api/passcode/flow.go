package passcode

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePasscodeConfirmation flowpilot.StateName = "passcode_confirmation"
)

const (
	ActionSubmitPasscode flowpilot.ActionName = "submit_email_passcode"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePasscodeConfirmation, SubmitPasscode{}).
	BeforeState(StatePasscodeConfirmation, shared.SendPasscode{}).
	MustBuild()
