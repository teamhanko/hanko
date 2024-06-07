package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailDelete struct {
	shared.Action
}

func (a EmailDelete) GetName() flowpilot.ActionName {
	return shared.ActionEmailDelete
}

func (a EmailDelete) GetDescription() string {
	return "Delete an email address."
}

func (a EmailDelete) Initialize(c flowpilot.InitializationContext) {
	if a.mustSuspend(c) {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("email_id").Required(true).Hidden(true))
}

func (a EmailDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	emailToBeDeletedId := uuid.FromStringOrNil(c.Input().Get("email_id").String())
	emailToBeDeletedModel := userModel.GetEmailById(emailToBeDeletedId)
	if emailToBeDeletedModel == nil {
		return c.ContinueFlowWithError(
			c.GetCurrentState(),
			flowpilot.ErrorFormDataInvalid.Wrap(errors.New("unknown email")),
		)
	}

	if emailToBeDeletedModel.IsPrimary() {
		if !deps.Cfg.Email.Optional {
			return c.ContinueFlowWithError(
				c.GetCurrentState(),
				flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("cannot delete primary email")),
			)
		} else {
			err := deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Delete(*emailToBeDeletedModel.PrimaryEmail)
			if err != nil {
				return fmt.Errorf("could not delete primary email: %w", err)
			}
		}
	}

	err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Delete(*emailToBeDeletedModel)
	if err != nil {
		return fmt.Errorf("could not delete email: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogEmailDeleted,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("email", emailToBeDeletedModel.Address),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	userModel.Emails.Delete(emailToBeDeletedId)

	return c.ContinueFlow(shared.StateProfileInit)
}

func (a EmailDelete) mustSuspend(c flowpilot.Context) bool {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return true
	}

	if len(userModel.Emails) == 0 {
		return true
	}

	isLastEmail := len(userModel.Emails) == 1
	canDoWebauthn := deps.Cfg.Passkey.Enabled && len(userModel.WebauthnCredentials) > 0
	canUseUsernameAsLoginIdentifier := deps.Cfg.Username.UseAsLoginIdentifier && userModel.Username.String != ""
	canUseEmailAsLoginIdentifier := deps.Cfg.Email.UseAsLoginIdentifier && !isLastEmail
	canDoPassword := deps.Cfg.Password.Enabled && userModel.PasswordCredential != nil && (canUseUsernameAsLoginIdentifier || canUseEmailAsLoginIdentifier)
	canDoThirdParty := deps.Cfg.ThirdParty.Providers.HasEnabled()
	canUseNoOtherAuthMethod := !canDoWebauthn && !canDoPassword && !canDoThirdParty

	if deps.Cfg.Email.Enabled && isLastEmail && (!deps.Cfg.Email.Optional || canUseNoOtherAuthMethod) {
		return true
	}

	return false
}
