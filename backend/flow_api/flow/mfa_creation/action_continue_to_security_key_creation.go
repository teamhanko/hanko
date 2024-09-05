package mfa_creation

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToSecurityKeyCreation struct {
	shared.Action
}

func (a ContinueToSecurityKeyCreation) GetName() flowpilot.ActionName {
	return shared.ActionContinueToSecurityKeyCreation
}

func (a ContinueToSecurityKeyCreation) GetDescription() string {
	return "Create a security key"
}

func (a ContinueToSecurityKeyCreation) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.MFA.Enabled || !deps.Cfg.MFA.SecurityKeys.Enabled {
		c.SuspendAction()
		return
	}

	webauthnAvailable := c.Stash().Get(shared.StashPathWebauthnAvailable).Bool()
	mustUsePlatformAttachment := deps.Cfg.MFA.SecurityKeys.AuthenticatorAttachment == "platform"
	platformAuthenticatorAvailable := c.Stash().Get(shared.StashPathWebauthnPlatformAuthenticatorAvailable).Bool()

	if !webauthnAvailable || (mustUsePlatformAttachment && !platformAuthenticatorAvailable) {
		c.SuspendAction()
	}
}

func (a ContinueToSecurityKeyCreation) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue(shared.StateMFASecurityKeyCreation)
}
