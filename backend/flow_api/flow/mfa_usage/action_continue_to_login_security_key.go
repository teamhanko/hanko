package mfa_usage

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToLoginSecurityKey struct {
	shared.Action
}

func (a ContinueToLoginSecurityKey) GetName() flowpilot.ActionName {
	return shared.ActionContinueToLoginSecurityKey
}

func (a ContinueToLoginSecurityKey) GetDescription() string {
	return "Continues the flow to the security key login."
}

func (a ContinueToLoginSecurityKey) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	attachmentSupported := c.Stash().Get(shared.StashPathSecurityKeyAttachmentSupported).Bool()

	if !deps.Cfg.MFA.SecurityKeys.Enabled || !c.Stash().Get(shared.StashPathUserHasWebauthnCredential).Bool() || !attachmentSupported {
		c.SuspendAction()
	}
}

func (a ContinueToLoginSecurityKey) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue(shared.StateLoginSecurityKey)
}
