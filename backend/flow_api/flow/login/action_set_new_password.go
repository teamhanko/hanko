package login

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	passkeyOnboarding "github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SetNewPassword struct {
	shared.Action
}

func (a SetNewPassword) GetName() flowpilot.ActionName {
	return ActionSetNewPassword
}

func (a SetNewPassword) GetDescription() string {
	return "Submit a new password."
}

func (a SetNewPassword) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	c.AddInputs(flowpilot.PasswordInput("new_password").Required(true).MinLength(deps.Cfg.Password.MinPasswordLength))
}

func (a SetNewPassword) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	newPassword := c.Input().Get("new_password").String()

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("user_id does not exist")))
	}

	authUserID := c.Stash().Get("user_id").String()

	err := deps.PasswordService.RecoverPassword(uuid.FromStringOrNil(authUserID), newPassword)
	if err != nil {
		if errors.Is(err, services.ErrorPasswordInvalid) {
			c.Input().SetError("password", flowpilot.ErrorValueInvalid)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid.Wrap(err))
		}

		return fmt.Errorf("could not recover password: %w", err)
	}

	// Decide which is the next state according to the config and user input
	if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passkeyOnboarding.StateIntroduction, StateSuccess)
	}

	return c.ContinueFlow(StateSuccess)
}
