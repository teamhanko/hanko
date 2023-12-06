package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
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

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(
			c.GetErrorState(),
			flowpilot.ErrorOperationNotPermitted.
				Wrap(errors.New("user_id does not exist")))
	}

	if !c.Stash().Get("primary_email").Exists() && !c.Stash().Get("username").Exists() {
		return errors.New("either email or username must exist in the stash")
	}

	userID, err := uuid.FromString(c.Stash().Get("user_id").String())
	if err != nil {
		return fmt.Errorf("failed to parse user id as a uuid: %w", err)
	}

	email := c.Stash().Get("primary_email").String()
	username := c.Stash().Get("username").String()

	params := services.GenerateCreationOptionsParams{
		Tx:       deps.Tx,
		UserID:   userID,
		Email:    email,
		Username: username,
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
