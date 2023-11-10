package flows

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/actions"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewCapabilitiesSubFlow(cfg config.Config) flowpilot.SubFlow {
	return flowpilot.NewSubFlow().
		State(common.StateLoginPreflight, actions.NewSendCapabilities(cfg)).
		MustBuild()
}
