package register_password

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePasswordCreation flowpilot.StateName = "password_creation"
)

const (
	ActionRegisterPassword flowpilot.ActionName = "register_password"
	ActionSkip             flowpilot.ActionName = "skip"
	ActionBack             flowpilot.ActionName = "back"
)

var SubFlow = flowpilot.NewSubFlow("register_password").
	State(StatePasswordCreation, RegisterPassword{}, Back{}, Skip{}).
	MustBuild()
