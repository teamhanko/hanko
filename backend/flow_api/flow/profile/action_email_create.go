package profile

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailCreate struct {
	shared.Action
}

func (a EmailCreate) GetName() flowpilot.ActionName {
	return ActionEmailCreate
}

func (a EmailCreate) GetDescription() string {
	return "Create an email address for the current session user."
}

func (a EmailCreate) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Identifier.Email.Enabled {
		c.SuspendAction()
	} else {
		c.AddInputs(flowpilot.EmailInput("email").Required(true))
	}
}

func (a EmailCreate) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	newEmailAddress := c.Input().Get("email").String()

	existingEmailModel, err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).FindByAddress(newEmailAddress)
	if err != nil {
		return fmt.Errorf("could not fetch email: %w", err)
	}

	if existingEmailModel != nil {
		if (existingEmailModel.UserID != nil && existingEmailModel.UserID.String() == userModel.ID.String()) || !deps.Cfg.Identifier.Email.Verification {
			c.Input().SetError("email", shared.ErrorEmailAlreadyExists)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		} else {
			err = c.CopyInputValuesToStash("email")
			if err != nil {
				return fmt.Errorf("failed to copy email to stash: %w", err)
			}

			err = c.Stash().Set("user_id", userModel.ID.String())
			if err != nil {
				return fmt.Errorf("failed to set user_id to stash: %w", err)
			}

			err = c.Stash().Set("passcode_template", "email_registration_attempted")
			if err != nil {
				return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
			}

			return c.StartSubFlow(passcode.StatePasscodeConfirmation)
		}
	} else if deps.Cfg.Identifier.Email.Verification {
		err = c.CopyInputValuesToStash("email")
		if err != nil {
			return fmt.Errorf("failed to copy email to stash: %w", err)
		}

		err = c.Stash().Set("user_id", userModel.ID.String())
		if err != nil {
			return fmt.Errorf("failed to set user_id to stash: %w", err)
		}

		err = c.Stash().Set("passcode_template", "email_verification")
		if err != nil {
			return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
		}

		return c.StartSubFlow(passcode.StatePasscodeConfirmation, StateProfileInit)
	} else {
		emailModel := models.NewEmail(&userModel.ID, newEmailAddress)

		err = deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Create(*emailModel)
		if err != nil {
			return fmt.Errorf("could not save email: %w", err)
		}

		emailModels, err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).FindByUserId(*emailModel.UserID)
		if err != nil {
			return fmt.Errorf("could fetch emails: %w", err)
		}

		if len(emailModels) == 1 && emailModels[0].ID.String() == emailModel.ID.String() {
			// The user has only one 1 email and it is the email we just added. It makes sense then,
			// to automatically set this as the primary email.
			primaryEmailModel := models.NewPrimaryEmail(emailModel.ID, userModel.ID)
			err = deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Create(*primaryEmailModel)
			if err != nil {
				return fmt.Errorf("could not save primary email: %w", err)
			}
		}

		err = deps.AuditLogger.CreateWithConnection(
			deps.Tx,
			deps.HttpContext,
			models.AuditLogEmailCreated,
			&models.User{ID: userModel.ID},
			nil,
			auditlog.Detail("email", emailModel.Address),
			auditlog.Detail("flow_id", c.GetFlowID()))

		if err != nil {
			return fmt.Errorf("could not create audit log: %w", err)
		}
		return c.ContinueFlow(StateProfileInit)
	}
}

func (a EmailCreate) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
