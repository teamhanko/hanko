package capabilities

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePreflight flowpilot.StateName = "preflight"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePreflight, SendCapabilities{}).
	MustBuild()
