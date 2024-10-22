package credential_onboarding

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnGenerateCreationOptions struct {
	shared.Action
}

func (a WebauthnGenerateCreationOptions) GetName() flowpilot.ActionName {
	return shared.ActionWebauthnGenerateCreationOptions
}

func (a WebauthnGenerateCreationOptions) GetDescription() string {
	return "Get creation options to create a webauthn credential."
}

func (a WebauthnGenerateCreationOptions) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get(shared.StashPathWebauthnAvailable).Bool() {
		c.SuspendAction()
	}
}

func (a WebauthnGenerateCreationOptions) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	if !c.Stash().Get(shared.StashPathUserID).Exists() {
		return errors.New("user_id does not exist in the stash")
	}

	if !c.Stash().Get(shared.StashPathEmail).Exists() && !c.Stash().Get(shared.StashPathUsername).Exists() {
		return errors.New("either email or username must exist in the stash")
	}

	userID, err := uuid.FromString(c.Stash().Get(shared.StashPathUserID).String())
	if err != nil {
		return fmt.Errorf("failed to parse user id as a uuid: %w", err)
	}

	email := c.Stash().Get(shared.StashPathEmail).String()
	username := c.Stash().Get(shared.StashPathUsername).String()

	params := services.GenerateCreationOptionsParams{
		Tx:       deps.Tx,
		UserID:   userID,
		Email:    &email,
		Username: &username,
	}

	sessionDataModel, creationOptions, err := deps.WebauthnService.GenerateCreationOptionsPasskey(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn creation options: %w", err)
	}

	err = c.Stash().Set(shared.StashPathWebauthnSessionDataID, sessionDataModel.ID)
	if err != nil {
		return err
	}

	err = c.Stash().Set(shared.StashPathCreateMFAOnlyCredential, false)
	if err != nil {
		return err
	}

	err = c.Payload().Set("creation_options", creationOptions)
	if err != nil {
		return err
	}

	return c.Continue(shared.StateOnboardingVerifyPasskeyAttestation)
}
