package states

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StateLoginInit             flowpilot.StateName = "login_init"
	StateLoginMethodChooser    flowpilot.StateName = "login_method_chooser"
	StateLoginPassword         flowpilot.StateName = "login_password"
	StateLoginPasskey          flowpilot.StateName = "login_passkey"
	StateLoginPasswordRecovery flowpilot.StateName = "login_password_recovery"
)
