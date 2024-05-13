package login_method_chooser

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/login_password"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateLoginMethodChooser flowpilot.StateName = "login_method_chooser"
)

const (
	ActionContinueToPasswordLogin        flowpilot.ActionName = "continue_to_password_login"
	ActionContinueToPasscodeConfirmation flowpilot.ActionName = "continue_to_passcode_confirmation"
)

var SubFlow = flowpilot.NewSubFlow("login_method_chooser").
	State(StateLoginMethodChooser,
		ContinueToPasswordLogin{},
		ContinueToPasscodeConfirmation{},
		shared.Back{},
	).
	SubFlows(login_password.SubFlow, passcode.SubFlow).
	MustBuild()
