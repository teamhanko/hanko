package capabilities

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePreflight flowpilot.StateName = "preflight"
)

const (
	ActionRegisterClientCapabilities flowpilot.ActionName = "register_client_capabilities"
)

var SubFlow = flowpilot.NewSubFlow("capabilities").
	State(StatePreflight, RegisterClientCapabilities{}).
	MustBuild()
