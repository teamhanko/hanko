package profile

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialCreate struct {
	shared.Action
}

func (a WebauthnCredentialCreate) GetName() flowpilot.ActionName {
	return ActionWebauthnCredentialCreate
}

func (a WebauthnCredentialCreate) GetDescription() string {
	return "Create a Webauthn credential for the current session user."
}

func (a WebauthnCredentialCreate) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("webauthn_available").Bool() {
		c.SuspendAction()
	}
}

func (a WebauthnCredentialCreate) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	primaryEmailModel := userModel.Emails.GetPrimary()
	if primaryEmailModel == nil && userModel.Username == "" {
		return errors.New("user must have either email or username")
	}

	var primaryEmailAddress string
	if primaryEmailModel != nil {
		primaryEmailAddress = primaryEmailModel.Address
	}

	params := services.GenerateCreationOptionsParams{
		Tx:       deps.Tx,
		UserID:   userModel.ID,
		Email:    primaryEmailAddress,
		Username: userModel.Username,
	}

	sessionDataModel, creationOptions, err := deps.WebauthnService.GenerateCreationOptions(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn creation options: %w", err)
	}

	err = c.Stash().Set("webauthn_session_data_id", sessionDataModel.ID)
	if err != nil {
		return err
	}

	err = c.Payload().Set("creation_options", creationOptions)
	if err != nil {
		return err
	}

	return c.ContinueFlow(StateProfileWebauthnCredentialVerification)
}

func (a WebauthnCredentialCreate) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
