package credential_onboarding

import (
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
	return c.EndSubFlow()
}

func (a SkipCredentialOnboardingMethodChooser) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
