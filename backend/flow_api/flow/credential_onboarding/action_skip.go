package credential_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"strings"
)

type Skip struct {
	shared.Action
}

func (a Skip) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a Skip) GetDescription() string {
	return "Skip"
}

func (a Skip) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	switch c.GetFlowName() {
	case "registration":
		if c.GetCurrentState() == shared.StatePasswordCreation {
			if !deps.Cfg.Password.Optional || !deps.Cfg.Email.RequireVerification {
				c.SuspendAction()
			}
		} else if c.GetCurrentState() == shared.StateOnboardingCreatePasskey {
			if !deps.Cfg.Passkey.Optional || !deps.Cfg.Email.RequireVerification {
				c.SuspendAction()
			}
		}

		if strings.Contains(c.GetFlowPath(), "registration_method_chooser") {
			c.SuspendAction()
		}
	}
}
func (a Skip) Execute(c flowpilot.ExecutionContext) error {
	switch c.GetFlowName() {
	case "registration":
		return c.EndSubFlow()
	case "login":
		if c.GetCurrentState() == shared.StatePasswordCreation {
			return c.ContinueFlow(shared.StateOnboardingCreatePasskey)
		} else {
			return c.ContinueFlow(shared.StatePasswordCreation)
		}
	}
	return nil
}

func (a Skip) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
