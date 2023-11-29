package preflight

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePreflight flowpilot.StateName = "preflight"
)

const (
	ActionRegisterClientCapabilities flowpilot.ActionName = "register_client_capabilities"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePreflight, RegisterClientCapabilities{}).
	MustBuild()
