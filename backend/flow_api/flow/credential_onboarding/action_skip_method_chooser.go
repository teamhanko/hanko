package credential_onboarding

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SkipCredentialOnboardingMethodChooser struct {
	shared.Action
}

func (a SkipCredentialOnboardingMethodChooser) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a SkipCredentialOnboardingMethodChooser) GetDescription() string {
	return "Skip"
}

func (a SkipCredentialOnboardingMethodChooser) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	exists := c.Stash().Get("email").Exists()
	if c.GetFlowName() == "registration" && !(deps.Cfg.Email.UseForAuthentication && exists) {
		c.SuspendAction()
	}
}

func (a SkipCredentialOnboardingMethodChooser) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set("suspend_back_action", false); err != nil {
		return fmt.Errorf("failed to set suspend_back_action to the stash: %w", err)
	}

	return c.EndSubFlow()
}
