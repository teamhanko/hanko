package credential_onboarding

import (
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
	switch c.GetFlowName() {
	case "registration":
		if !deps.Cfg.Password.Optional || !deps.Cfg.Email.RequireVerification {
			c.SuspendAction()
		}

		if c.GetFlowPath().HasFragment("registration_method_chooser") {
			c.SuspendAction()
		}
	}
}
func (a SkipPassword) Execute(c flowpilot.ExecutionContext) error {
	if prevState, _ := c.GetPreviousState(); prevState != nil && *prevState == shared.StateOnboardingVerifyPasskeyAttestation {
		return c.EndSubFlow()
	}

	return c.ContinueFlow(shared.StateOnboardingCreatePasskey)

}

func (a SkipPassword) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
