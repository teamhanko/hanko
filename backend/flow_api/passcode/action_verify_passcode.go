package passcode

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type VerifyPasscode struct {
	shared.Action
}

func (a VerifyPasscode) GetName() flowpilot.ActionName {
	return ActionVerifyPasscode
}

func (a VerifyPasscode) GetDescription() string {
	return "Enter a passcode."
}

func (a VerifyPasscode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("code").Required(true))
}

func (a VerifyPasscode) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	passcodeId, err := uuid.FromString(c.Stash().Get("passcode_id").String())
	if err != nil {
		return err
	}

	err = deps.PasscodeService.VerifyPasscode(deps.Tx, passcodeId, c.Input().Get("code").String())
	if err != nil {
		if errors.Is(err, services.ErrorPasscodeInvalid) ||
			errors.Is(err, services.ErrorPasscodeNotFound) ||
			errors.Is(err, services.ErrorPasscodeExpired) {
			return c.ContinueFlowWithError(c.GetErrorState(), shared.ErrorPasscodeInvalid)
		}

		if errors.Is(err, services.ErrorPasscodeMaxAttemptsReached) {
			return c.ContinueFlowWithError(c.GetErrorState(), shared.ErrorPasscodeMaxAttemptsReached)
		}

		return fmt.Errorf("failed to verify passcode: %w", err)
	}

	err = c.Stash().Delete("passcode_id")
	if err != nil {
		return fmt.Errorf("failed to delete passcode_id from stash: %w", err)
	}

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("account does not exist")))
	}

	err = c.Stash().Set("email_verified", true) // TODO: maybe change attribute path
	if err != nil {
		return err
	}

	return c.EndSubFlow()
}
