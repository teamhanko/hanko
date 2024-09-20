package profile

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SecurityKeyCreate struct {
	shared.Action
}

func (a SecurityKeyCreate) GetName() flowpilot.ActionName {
	return shared.ActionSecurityKeyCreate
}

func (a SecurityKeyCreate) GetDescription() string {
	return "Get WebAuthn creation options to register a WebAuthn credential."
}

func (a SecurityKeyCreate) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if !deps.Cfg.MFA.Enabled || !deps.Cfg.MFA.SecurityKeys.Enabled {
		c.SuspendAction()
	}

	if !c.Stash().Get(shared.StashPathWebauthnAvailable).Bool() {
		c.SuspendAction()
	}

	if len(userModel.GetSecurityKeys()) >= deps.Cfg.MFA.SecurityKeys.Limit {
		c.SuspendAction()
		return
	}
}

func (a SecurityKeyCreate) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

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
		Email:    &primaryEmailAddress,
		Username: userModel.GetUsername(),
	}

	sessionDataModel, creationOptions, err := deps.WebauthnService.GenerateCreationOptionsSecurityKey(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn creation options: %w", err)
	}

	err = c.Stash().Set(shared.StashPathWebauthnSessionDataID, sessionDataModel.ID)
	if err != nil {
		return err
	}

	err = c.Stash().Set(shared.StashPathCreateMFAOnlyCredential, true)
	if err != nil {
		return err
	}

	err = c.Payload().Set("creation_options", creationOptions)
	if err != nil {
		return err
	}

	return c.Continue(shared.StateProfileWebauthnCredentialVerification)
}
