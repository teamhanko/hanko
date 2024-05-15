package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/register_password"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type Skip struct {
	shared.Action
}

func (a Skip) GetName() flowpilot.ActionName {
	return ActionSkip
}

func (a Skip) GetDescription() string {
	return "Skip the passkey onboarding"
}

func (a Skip) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	switch c.GetFlowName() {
	case "registration":
		if !deps.Cfg.Passkey.Optional || !deps.Cfg.Email.RequireVerification {
			// Skip is only available when passkeys are optional or the email has been verified beforehand to ensure the
			// user has a credential after registration.
			c.SuspendAction()
		}
	case "login":
		if !deps.Cfg.Passkey.Optional {
			c.SuspendAction()
		}
	}
}

func (a Skip) Execute(c flowpilot.ExecutionContext) error {
	switch c.GetFlowName() {
	case "registration":
		return c.EndSubFlow()
	case "login":
		return c.StartSubFlow(register_password.StatePasswordCreation)
	}

	return nil
}

func (a Skip) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
