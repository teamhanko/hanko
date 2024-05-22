package credential_onboarding

import (
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
	switch c.GetFlowName() {
	case "registration":
		if !deps.Cfg.Passkey.Optional || !deps.Cfg.Email.RequireVerification {
			c.SuspendAction()
		}

		if c.GetFlowPath().HasFragment("registration_method_chooser") {
			c.SuspendAction()
		}
	}
}
func (a SkipPasskey) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if deps.Cfg.Password.AcquireOnLogin == "conditional" && !c.Stash().Get("user_has_password").Bool() {
		return c.ContinueFlow(shared.StatePasswordCreation)
	}

	return c.EndSubFlow()

}

func (a SkipPasskey) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
