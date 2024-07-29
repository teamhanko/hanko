package profile

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialDelete struct {
	shared.Action
}

func (a WebauthnCredentialDelete) GetName() flowpilot.ActionName {
	return shared.ActionWebauthnCredentialDelete
}

func (a WebauthnCredentialDelete) GetDescription() string {
	return "Delete a Webauthn credential."
}

func (a WebauthnCredentialDelete) Initialize(c flowpilot.InitializationContext) {
	if a.mustSuspend(c) {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("passkey_id").Required(true).Hidden(true))
}

func (a WebauthnCredentialDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	webauthnCredentialModel := userModel.GetWebauthnCredentialById(c.Input().Get("passkey_id").String())
	if webauthnCredentialModel == nil {
		return c.Error(shared.ErrorNotFound)
	}

	err := deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Delete(*webauthnCredentialModel)
	if err != nil {
		return fmt.Errorf("could not delete passkey: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogPasskeyDeleted,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("credential_id", webauthnCredentialModel.ID),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	userModel.DeleteWebauthnCredential(webauthnCredentialModel.ID)

	return c.Continue(shared.StateProfileInit)
}

func (a WebauthnCredentialDelete) mustSuspend(c flowpilot.Context) bool {
	deps := a.GetDeps(c)

	if !deps.Cfg.Passkey.Enabled {
		return true
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return true
	}

	if len(userModel.WebauthnCredentials) == 0 {
		return true
	}

	identities := userModel.GetIdentities()

	isLastWebauthnCredential := len(userModel.WebauthnCredentials) == 1

	if isLastWebauthnCredential && !deps.Cfg.Passkey.Optional {
		return true
	}

	canUseUsernameAsLoginIdentifier := deps.Cfg.Username.UseAsLoginIdentifier && userModel.Username != nil
	canUseEmailAsLoginIdentifier := deps.Cfg.Email.UseAsLoginIdentifier && len(userModel.Emails) > 0
	canDoPassword := deps.Cfg.Password.Enabled && userModel.PasswordCredential != nil && (canUseUsernameAsLoginIdentifier || canUseEmailAsLoginIdentifier)
	canDoPasscode := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseForAuthentication && (canUseEmailAsLoginIdentifier || canUseUsernameAsLoginIdentifier && len(userModel.Emails) > 0)
	canDoThirdParty := services.UserCanDoThirdParty(deps.Cfg, identities) || services.UserCanDoSaml(deps.Cfg, identities)
	canUseNoOtherAuthMethod := !canDoPassword && !canDoThirdParty && !canDoPasscode

	if isLastWebauthnCredential && canUseNoOtherAuthMethod {
		return true
	}

	return false
}
