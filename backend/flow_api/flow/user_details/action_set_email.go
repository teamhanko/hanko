package user_details

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
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
		MinLength(3).
		MaxLength(255))
}

func (a EmailAddressSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userID := uuid.FromStringOrNil(c.Stash().Get("user_id").String())
	user, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userID)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user does not exists (id: %s)", userID.String())
	}

	email := c.Input().Get("email").String()
	emailModel := models.NewEmail(&userID, email)

	err = deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Create(*emailModel)
	if err != nil {
		return fmt.Errorf("failed to create a new email: %w", err)
	}

	primaryEmail := models.NewPrimaryEmail(emailModel.ID, userID)
	err = deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Create(*primaryEmail)
	if err != nil {
		return fmt.Errorf("failed to create a new primary email: %w", err)
	}

	err = c.Stash().Set("email", email)
	if err != nil {
		return fmt.Errorf("failed to set email to the stash: %w", err)
	}

	return c.EndSubFlow()
}

func (a EmailAddressSet) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
