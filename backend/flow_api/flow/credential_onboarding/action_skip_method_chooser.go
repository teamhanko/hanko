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
	emailExists := c.Stash().Get(shared.StashPathEmail).Exists()

	if c.GetFlowName() == shared.FlowRegistration &&
		!(deps.Cfg.Email.UseForAuthentication && emailExists) {
		c.SuspendAction()
	}
}

func (a SkipCredentialOnboardingMethodChooser) Execute(c flowpilot.ExecutionContext) error {
	if err := c.DeleteStateHistory(true); err != nil {
		return fmt.Errorf("failed to delete the state history: %w", err)
	}

	return c.EndSubFlow()
}
