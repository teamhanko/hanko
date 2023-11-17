package capabilities

import (
	"github.com/teamhanko/hanko/backend/flow_api/capabilities/actions"
	"github.com/teamhanko/hanko/backend/flow_api/capabilities/states"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

var SubFlow = flowpilot.NewSubFlow().
	State(states.StatePreflight, actions.SendCapabilities{}).
	MustBuild()
