package credential_onboarding

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SkipPassword struct {
	shared.Action
}

func (a SkipPassword) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a SkipPassword) GetDescription() string {
	return "Skip"
}

func (a SkipPassword) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Optional {
		c.SuspendAction()
	}

	if prev, _ := c.GetPreviousState(); prev != nil && *prev == shared.StateCredentialOnboardingChooser {
		c.SuspendAction()
	}

}
func (a SkipPassword) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if err := c.Stash().Set("suspend_back_action", false); err != nil {
		return fmt.Errorf("failed to set suspend_back_action to the stash: %w", err)
	}

	if c.GetFlowName() == "login" {
		if deps.Cfg.Passkey.Enabled && deps.Cfg.Passkey.AcquireOnLogin == "conditional" && !c.Stash().Get("user_has_webauthn_credential").Bool() && c.Stash().Get("webauthn_available").Bool() {
			return c.ContinueFlow(shared.StateOnboardingCreatePasskey)
		}
	} else if c.GetFlowName() == "registration" {
		if deps.Cfg.Passkey.Enabled && deps.Cfg.Passkey.AcquireOnRegistration == "conditional" && !c.Stash().Get("user_has_webauthn_credential").Bool() && c.Stash().Get("webauthn_available").Bool() {
			return c.ContinueFlow(shared.StateOnboardingCreatePasskey)
		}
	}

	return c.EndSubFlow()
}
