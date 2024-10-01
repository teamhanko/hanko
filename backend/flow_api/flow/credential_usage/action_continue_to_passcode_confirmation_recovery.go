package credential_usage

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmationRecovery struct {
	shared.Action
}

func (a ContinueToPasscodeConfirmationRecovery) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasscodeConfirmationRecovery
}

func (a ContinueToPasscodeConfirmationRecovery) GetDescription() string {
	return "Send a recovery passcode code via email."
}

func (a ContinueToPasscodeConfirmationRecovery) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Recovery || len(c.Stash().Get(shared.StashPathEmail).String()) == 0 {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmationRecovery) Execute(c flowpilot.ExecutionContext) error {
	if len(c.Stash().Get(shared.StashPathUserID).String()) > 0 {
		if err := c.Stash().Set(shared.StashPathPasscodeTemplate, "recovery"); err != nil {
			return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
		}
	} else {
		if err := c.Stash().Set(shared.StashPathPasscodeTemplate, "email_login_attempted"); err != nil {
			return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
		}
	}

	_ = c.Stash().Set(shared.StashPathPasswordRecoveryPending, true)

	return c.Continue(shared.StatePasscodeConfirmation)
}
