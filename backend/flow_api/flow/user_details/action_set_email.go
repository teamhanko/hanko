package user_details

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"strings"
)

type EmailAddressSet struct {
	shared.Action
}

func (a EmailAddressSet) GetName() flowpilot.ActionName {
	return shared.ActionEmailAddressSet
}

func (a EmailAddressSet) GetDescription() string {
	return "Set a new email address."
}

func (a EmailAddressSet) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.StringInput("email").
		Required(!deps.Cfg.Email.Optional).
		MaxLength(deps.Cfg.Email.MaxLength))
}

func (a EmailAddressSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userID := uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())
	user, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userID)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user does not exists (id: %s)", userID.String())
	}

	email := strings.TrimSpace(c.Input().Get("email").String())
	emailModel := models.NewEmail(&userID, email)

	existingEmail, err := deps.Persister.GetEmailPersister().FindByAddress(email)
	if err != nil {
		return fmt.Errorf("failed to get email from db: %w", err)
	}

	if existingEmail != nil {
		c.Input().SetError("email", shared.ErrorEmailAlreadyExists)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	err = deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Create(*emailModel)
	if err != nil {
		return fmt.Errorf("failed to create a new email: %w", err)
	}

	primaryEmail := models.NewPrimaryEmail(emailModel.ID, userID)
	err = deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Create(*primaryEmail)
	if err != nil {
		return fmt.Errorf("failed to create a new primary email: %w", err)
	}

	err = c.Stash().Set(shared.StashPathEmail, email)
	if err != nil {
		return fmt.Errorf("failed to set email to the stash: %w", err)
	}

	err = c.Stash().Set(shared.StashPathPasscodeTemplate, "email_verification")
	if err != nil {
		return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
	}

	if deps.Cfg.Email.RequireVerification {
		return c.Continue(shared.StatePasscodeConfirmation)
	}

	err = c.DeleteStateHistory(true)
	if err != nil {
		return fmt.Errorf("failed to delete the state history: %w", err)
	}

	return c.Continue()
}
