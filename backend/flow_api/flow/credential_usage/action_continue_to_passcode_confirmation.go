package credential_usage

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
)

type ContinueToPasscodeConfirmation struct {
	shared.Action
}

func (a ContinueToPasscodeConfirmation) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasscodeConfirmation
}

func (a ContinueToPasscodeConfirmation) GetDescription() string {
	return "Send a login passcode code via email."
}

func (a ContinueToPasscodeConfirmation) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	if deps.Cfg.Privacy.OnlyShowActualLoginMethods && (!c.Stash().Get(shared.StashPathUserHasEmails).Bool() || !deps.Cfg.Email.Enabled || (deps.Cfg.Email.Enabled && !deps.Cfg.Email.UseForAuthentication)) {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmation) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set(shared.StashPathLoginMethod, "passcode"); err != nil {
		return fmt.Errorf("failed to set login_method to stash: %w", err)
	}

	if len(c.Stash().Get(shared.StashPathUserID).String()) > 0 {
		if err := c.Stash().Set(shared.StashPathPasscodeTemplate, shared.PasscodeTemplateLogin); err != nil {
			return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
		}
	} else {
		if err := c.Stash().Set(shared.StashPathPasscodeTemplate, shared.PasscodeTemplateEmailLoginAttempted); err != nil {
			return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
		}
	}

	return c.Continue(shared.StatePasscodeConfirmation)
}
