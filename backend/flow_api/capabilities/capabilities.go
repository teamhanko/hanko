package capabilities

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/capabilities/actions"
	"github.com/teamhanko/hanko/backend/flow_api/capabilities/states"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewCapabilitiesSubFlow(cfg config.Config) flowpilot.SubFlow {
	return flowpilot.NewSubFlow().
		State(states.StatePreflight, actions.NewSendCapabilities(cfg)).
		MustBuild()
}
