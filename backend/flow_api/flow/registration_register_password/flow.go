package registration_register_password

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePasswordCreation flowpilot.StateName = "password_creation"
)

const (
	ActionRegisterPassword flowpilot.ActionName = "register_password"
	ActionSkip             flowpilot.ActionName = "skip"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePasswordCreation, RegisterPassword{}, shared.Back{}, Skip{}).
	MustBuild()
