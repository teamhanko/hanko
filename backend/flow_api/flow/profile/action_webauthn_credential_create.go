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
	return shared.ActionWebauthnCredentialCreate
}

func (a WebauthnCredentialCreate) GetDescription() string {
	return "Create a Webauthn credential for the current session user."
}

func (a WebauthnCredentialCreate) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)

	if !deps.Cfg.Passkey.Enabled || !c.Stash().Get(shared.StashPathWebauthnAvailable).Bool() || (ok && len(userModel.WebauthnCredentials) >= deps.Cfg.Passkey.Limit) {
		c.SuspendAction()
	}
}

func (a WebauthnCredentialCreate) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	primaryEmailModel := userModel.Emails.GetPrimary()
	if primaryEmailModel == nil && userModel.Username == nil {
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
		Username: userModel.GetUsername(),
	}

	sessionDataModel, creationOptions, err := deps.WebauthnService.GenerateCreationOptions(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn creation options: %w", err)
	}

	err = c.Stash().Set(shared.StashPathWebauthnSessionDataID, sessionDataModel.ID)
	if err != nil {
		return err
	}

	err = c.Payload().Set("creation_options", creationOptions)
	if err != nil {
		return err
	}

	return c.Continue(shared.StateProfileWebauthnCredentialVerification)
}
