package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"golang.org/x/exp/slices"
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
	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	input := flowpilot.StringInput("email_id").Required(true).Hidden(true)

	if !a.emailDeletionAllowed(c, userModel) {
		c.SuspendAction()
		return
	}

	lastEmail := len(userModel.Emails) == 1
	deletableEmails := make(models.Emails, len(userModel.Emails))

	copy(deletableEmails, userModel.Emails)

	slices.DeleteFunc(deletableEmails, func(email models.Email) bool {
		return email.IsPrimary() && !lastEmail
	})

	for _, email := range deletableEmails {
		input.AllowedValue(email.Address, email.ID.String())
	}

	c.AddInputs(input)
}

func (a EmailDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	emailToBeDeletedId := uuid.FromStringOrNil(c.Input().Get("email_id").String())
	emailToBeDeletedModel := userModel.GetEmailById(emailToBeDeletedId)
	if emailToBeDeletedModel == nil {
		return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(errors.New("unknown email")))
	}

	if emailToBeDeletedModel.IsPrimary() {
		if !deps.Cfg.Email.Optional {
			return c.Error(flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("cannot delete primary email")))
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

	userModel.DeleteEmail(*emailToBeDeletedModel)

	return c.Continue(shared.StateProfileInit)
}

func (a EmailDelete) emailDeletionAllowed(c flowpilot.Context, userModel *models.User) bool {
	deps := a.GetDeps(c)

	if len(userModel.Emails) == 0 {
		return false
	}

	isLastEmail := len(userModel.Emails) == 1
	canDoWebauthn := deps.Cfg.Passkey.Enabled && len(userModel.WebauthnCredentials) > 0
	canUseUsernameAsLoginIdentifier := deps.Cfg.Username.UseAsLoginIdentifier && userModel.Username.String != ""
	canUseEmailAsLoginIdentifier := deps.Cfg.Email.UseAsLoginIdentifier && !isLastEmail
	canDoPassword := deps.Cfg.Password.Enabled && userModel.PasswordCredential != nil && (canUseUsernameAsLoginIdentifier || canUseEmailAsLoginIdentifier)
	canDoThirdParty := deps.Cfg.ThirdParty.Providers.HasEnabled() || (deps.Cfg.Saml.Enabled && len(deps.SamlService.Providers()) > 0)
	canUseNoOtherAuthMethod := !canDoWebauthn && !canDoPassword && !canDoThirdParty

	if deps.Cfg.Email.Enabled && isLastEmail && (!deps.Cfg.Email.Optional || canUseNoOtherAuthMethod) {
		return false
	}

	return true
}
