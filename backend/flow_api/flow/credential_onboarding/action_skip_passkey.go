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
	emailExists := c.Stash().Get(shared.StashPathEmail).Exists()
	canLoginWithEmail := emailExists && deps.Cfg.Email.Enabled && deps.Cfg.Email.UseForAuthentication

	if !deps.Cfg.Passkey.Optional {
		c.SuspendAction()
	}

	if c.IsPreviousState(shared.StateCredentialOnboardingChooser) {
		c.SuspendAction()
	}

	if c.IsPreviousState(shared.StatePasswordCreation) &&
		!c.Stash().Get(shared.StashPathUserHasPassword).Bool() &&
		!canLoginWithEmail {
		c.SuspendAction()
	}

	if c.IsPreviousState(shared.StatePasscodeConfirmation) &&
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
		!c.Stash().Get(shared.StashPathUserHasPassword).Bool() {
		return c.Continue(shared.StatePasswordCreation)
	}

	return c.Continue()
}

func (a SkipPasskey) acquirePassword(c flowpilot.Context, acquireType string) bool {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Enabled {
		return false
	}

	if c.IsFlow(shared.FlowLogin) && deps.Cfg.Password.AcquireOnLogin == acquireType {
		return true
	}

	if c.IsFlow(shared.FlowRegistration) && deps.Cfg.Password.AcquireOnRegistration == acquireType {
		return true
	}

	return false
}
