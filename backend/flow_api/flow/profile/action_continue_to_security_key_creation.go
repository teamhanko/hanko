package profile

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
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
		return
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if len(userModel.GetSecurityKeys()) >= deps.Cfg.MFA.SecurityKeys.Limit {
		c.SuspendAction()
	}
}

func (a ContinueToSecurityKeyCreation) Execute(c flowpilot.ExecutionContext) error {
	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	_ = c.Stash().Set(shared.StashPathUserID, userModel.ID)
	_ = c.Stash().Set(shared.StashPathEmail, userModel.Emails.GetPrimary().Address)
	_ = c.Stash().Set(shared.StashPathUsername, userModel.Username)

	return c.Continue(shared.StateMFASecurityKeyCreation, shared.StateProfileInit)
}
