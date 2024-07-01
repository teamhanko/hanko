package credential_onboarding

import (
	"fmt"
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
	emailExists := c.Stash().Get(shared.StashPathEmail).Exists()
	canLoginWithEmail := emailExists &&
		deps.Cfg.Email.Enabled &&
		deps.Cfg.Email.UseForAuthentication &&
		deps.Cfg.Email.UseAsLoginIdentifier

	if !deps.Cfg.Password.Optional {
		c.SuspendAction()
	}

	if c.IsPreviousState(shared.StateCredentialOnboardingChooser) {
		c.SuspendAction()
	}

	if c.IsPreviousState(shared.StateOnboardingCreatePasskey) &&
		!c.Stash().Get(shared.StashPathUserHasWebauthnCredential).Bool() &&
		!canLoginWithEmail {
		c.SuspendAction()
	}

	if (c.IsPreviousState(shared.StatePasscodeConfirmation) || c.IsPreviousState(shared.StateRegistrationInit)) &&
		!a.acquirePasskey(c, "always") &&
		!canLoginWithEmail {
		c.SuspendAction()
	}
}

func (a SkipPassword) Execute(c flowpilot.ExecutionContext) error {
	if !c.IsFlow(shared.FlowRegistration) {
		if err := c.DeleteStateHistory(true); err != nil {
			return fmt.Errorf("failed to delete state history: %w", err)
		}
	}

	if a.acquirePasskey(c, "conditional") &&
		!c.Stash().Get(shared.StashPathUserHasWebauthnCredential).Bool() &&
		c.Stash().Get(shared.StashPathWebauthnAvailable).Bool() {
		return c.Continue(shared.StateOnboardingCreatePasskey)
	}

	return c.Continue()
}

func (a SkipPassword) acquirePasskey(c flowpilot.Context, acquireType string) bool {
	deps := a.GetDeps(c)

	if !deps.Cfg.Passkey.Enabled {
		return false
	}

	if c.IsFlow(shared.FlowLogin) && deps.Cfg.Passkey.AcquireOnLogin == acquireType {
		return true
	}

	if c.IsFlow(shared.FlowRegistration) && deps.Cfg.Passkey.AcquireOnRegistration == acquireType {
		return true
	}

	return false
}
