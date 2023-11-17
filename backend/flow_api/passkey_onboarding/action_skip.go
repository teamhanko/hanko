package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type Skip struct {
	cfg config.Config
}

func (m Skip) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (m Skip) GetDescription() string {
	return "Skip the passkey onboarding"
}

func (m Skip) Initialize(c flowpilot.InitializationContext) {
	if !m.cfg.Passcode.Enabled && !m.cfg.Password.Enabled {
		// suspend action when only passkeys are allowed
		c.SuspendAction()
	}
}

func (m Skip) Execute(c flowpilot.ExecutionContext) error {
	return c.EndSubFlow()
}
