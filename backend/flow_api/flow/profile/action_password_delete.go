package profile

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasswordDelete struct {
	shared.Action
}

func (a PasswordDelete) GetName() flowpilot.ActionName {
	return shared.ActionPasswordDelete
}

func (a PasswordDelete) GetDescription() string {
	return "Delete a password."
}

func (a PasswordDelete) Initialize(c flowpilot.InitializationContext) {
	if a.mustSuspend(c) {
		c.SuspendAction()
		return
	}
}

func (a PasswordDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	passwordCredentialModel, err := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).GetByUserID(userModel.ID)
	if err != nil {
		return fmt.Errorf("could not fetch password credential: %w", err)
	}

	if passwordCredentialModel == nil {
		return c.Continue(shared.StateProfileInit)
	}

	err = deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).Delete(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("could not delete password credential: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogPasswordDeleted,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	userModel.PasswordCredential = nil

	return c.Continue(shared.StateProfileInit)
}

func (a PasswordDelete) mustSuspend(c flowpilot.Context) bool {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Enabled {
		return true
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return true
	}

	if userModel.PasswordCredential == nil {
		return true
	}

	canDoWebauthn := deps.Cfg.Passkey.Enabled && len(userModel.WebauthnCredentials) > 0
	canUseUsernameAsLoginIdentifier := deps.Cfg.Username.UseAsLoginIdentifier && userModel.Username.String != ""
	canUseEmailAsLoginIdentifier := deps.Cfg.Email.UseAsLoginIdentifier && len(userModel.Emails) > 0
	canDoPasscode := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseForAuthentication && (canUseEmailAsLoginIdentifier || canUseUsernameAsLoginIdentifier && len(userModel.Emails) > 0)
	canDoThirdParty := deps.Cfg.ThirdParty.Providers.HasEnabled() || (deps.Cfg.Saml.Enabled && len(deps.SamlService.Providers()) > 0)
	canUseNoOtherAuthMethod := !canDoWebauthn && !canDoPasscode && !canDoThirdParty

	if canUseNoOtherAuthMethod {
		return true
	}

	return false
}
