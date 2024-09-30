package credential_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/registration"
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

	if !deps.Cfg.Password.Optional && !deps.Cfg.Passkey.Optional {
		c.SuspendAction()
	}
}

func (a SkipCredentialOnboardingMethodChooser) Execute(c flowpilot.ExecutionContext) error {
	c.PreventRevert()

	if err := c.ExecuteHook(registration.ScheduleMFACreationStates{}); err != nil {
		return err
	}

	return c.Continue()
}
