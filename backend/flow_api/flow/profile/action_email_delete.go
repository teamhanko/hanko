package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
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
	deps := a.GetDeps(c)
	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	input := flowpilot.StringInput("email_id").Required(true).Hidden(true)

	lastEmail := len(userModel.Emails) == 1

	canDoPasskeyLogin := deps.Cfg.Passkey.Enabled && len(userModel.GetPasskeys()) > 0
	canDoPWLogin := deps.Cfg.Password.Enabled && userModel.PasswordCredential != nil
	canDoPasscode := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseForAuthentication

	for _, email := range userModel.Emails {
		if email.IsPrimary() {
			canDoPWLoginWithUsername := canDoPWLogin && deps.Cfg.Username.UseAsLoginIdentifier && userModel.GetUsername() != nil
			if lastEmail && deps.Cfg.Email.Optional && (canDoPasskeyLogin || canDoPWLoginWithUsername) {
				input.AllowedValue(email.Address, email.ID.String())
			}
		} else {
			if !canDoPasskeyLogin && !canDoPWLogin && !canDoPasscode {
				for _, otherEmail := range userModel.Emails {
					if otherEmail.ID.String() == email.ID.String() {
						continue
					}

					if services.UserCanDoThirdParty(deps.Cfg, otherEmail.Identities) ||
						services.UserCanDoSaml(deps.Cfg, otherEmail.Identities) {
						input.AllowedValue(email.Address, email.ID.String())
						break
					}
				}
			} else {
				input.AllowedValue(email.Address, email.ID.String())
			}
		}
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

	utils.NotifyUserChange(deps.HttpContext, deps.Tx, deps.Persister, events.UserEmailDelete, userModel.ID)

	return c.Continue(shared.StateProfileInit)
}
