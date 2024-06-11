package credential_onboarding

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SkipPasskey struct {
	shared.Action
}

func (a SkipPasskey) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a SkipPasskey) GetDescription() string {
	return "Skip"
}

func (a SkipPasskey) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Passkey.Optional {
		c.SuspendAction()
	}

	if prev, _ := c.GetPreviousState(); prev != nil && *prev == shared.StateCredentialOnboardingChooser {
		c.SuspendAction()
	}

}
func (a SkipPasskey) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if err := c.Stash().Set("suspend_back_action", false); err != nil {
		return fmt.Errorf("failed to set suspend_back_action to the stash: %w", err)
	}

	if c.GetFlowName() == "login" {
		if deps.Cfg.Password.Enabled && deps.Cfg.Password.AcquireOnLogin == "conditional" && !c.Stash().Get("user_has_password").Bool() {
			return c.ContinueFlow(shared.StatePasswordCreation)
		}
	} else if c.GetFlowName() == "registration" {
		if deps.Cfg.Password.Enabled && deps.Cfg.Password.AcquireOnRegistration == "conditional" && !c.Stash().Get("user_has_password").Bool() {
			return c.ContinueFlow(shared.StatePasswordCreation)
		}
	}

	return c.EndSubFlow()

}
