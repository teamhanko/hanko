package actions

import (
	"errors"
	"github.com/teamhanko/hanko/backend/config"
	passkeyOnboardingStates "github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/states"

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
	return "Skip the current state."
}

func (m Skip) Initialize(c flowpilot.InitializationContext) {
	if !m.cfg.Passcode.Enabled && !m.cfg.Password.Enabled {
		// suspend action when only passkeys are allowed
		c.SuspendAction()
	}
}

func (m Skip) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	switch c.GetCurrentState() {
	case passkeyOnboardingStates.StateOnboardingCreatePasskey:
		return c.EndSubFlow()
	default:
		// return an error, so we don't implicitly continue to unwanted state
		return errors.New("no destination is defined")
	}
}
