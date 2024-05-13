package login_password

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateLoginPassword         flowpilot.StateName = "login_password"
	StateLoginPasswordRecovery flowpilot.StateName = "login_password_recovery"
)

const (
	ActionContinueToPasscodeConfirmationRecovery flowpilot.ActionName = "continue_to_passcode_confirmation_recovery"
	ActionPasswordLogin                          flowpilot.ActionName = "password_login"
	ActionPasswordRecovery                       flowpilot.ActionName = "password_recovery"
)

var SubFlow = flowpilot.NewSubFlow("login_password").
	State(StateLoginPassword,
		PasswordLogin{},
		ContinueToPasscodeConfirmationRecovery{},
		shared.Back{},
	).
	State(StateLoginPasswordRecovery, PasswordRecovery{}).
	SubFlows(passcode.SubFlow).
	MustBuild()
