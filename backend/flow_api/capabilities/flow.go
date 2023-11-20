package capabilities

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePreflight flowpilot.StateName = "preflight"
)

const (
	ActionSendCapabilities flowpilot.ActionName = "send_capabilities"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePreflight, SendCapabilities{}).
	MustBuild()
