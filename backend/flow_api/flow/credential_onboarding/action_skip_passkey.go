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

	if c.GetPreviousState() == shared.StateCredentialOnboardingChooser {
		c.SuspendAction()
	}

	emailExists := c.Stash().Get("email").Exists()
	canLoginWithEmail := emailExists && deps.Cfg.Email.Enabled && deps.Cfg.Email.UseForAuthentication

	if c.GetPreviousState() == shared.StatePasswordCreation &&
		!c.Stash().Get("user_has_password").Bool() &&
		!canLoginWithEmail {
		c.SuspendAction()
	}

	if c.GetPreviousState() == shared.StatePasscodeConfirmation &&
		!a.acquirePassword(c, "always") &&
		!canLoginWithEmail {
		c.SuspendAction()
	}
}
func (a SkipPasskey) Execute(c flowpilot.ExecutionContext) error {
	if err := c.DeleteStateHistory(true); err != nil {
		return fmt.Errorf("failed to delete the state history: %w", err)
	}

	if a.acquirePassword(c, "conditional") &&
		!c.Stash().Get("user_has_password").Bool() {
		return c.ContinueFlow(shared.StatePasswordCreation)
	}

	return c.EndSubFlow()
}

func (a SkipPasskey) acquirePassword(c flowpilot.Context, acquireType string) bool {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Enabled {
		return false
	}

	if c.GetFlowName() == "login" && deps.Cfg.Password.AcquireOnLogin == acquireType {
		return true
	}

	if c.GetFlowName() == "registration" && deps.Cfg.Password.AcquireOnRegistration == acquireType {
		return true
	}

	return false
}
