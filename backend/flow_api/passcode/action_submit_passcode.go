package passcode

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var maxPasscodeTries = 3

type SubmitPasscode struct {
	shared.Action
}

func (a SubmitPasscode) GetName() flowpilot.ActionName {
	return ActionSubmitPasscode
}

func (a SubmitPasscode) GetDescription() string {
	return "Enter a passcode."
}

func (a SubmitPasscode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("code").Required(true))
}

func (a SubmitPasscode) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	passcodeId, err := uuid.FromString(c.Stash().Get("passcode_id").String())
	if err != nil {
		return err
	}

	passcode, err := deps.Persister.GetPasscodePersister().Get(passcodeId)
	if err != nil {
		return err
	}
	if passcode == nil {
		return errors.New("passcode not found")
	}

	expirationTime := passcode.CreatedAt.Add(time.Duration(passcode.Ttl) * time.Second)
	if expirationTime.Before(time.Now().UTC()) {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid.Wrap(errors.New("passcode is expired")))
	}

	err = bcrypt.CompareHashAndPassword([]byte(passcode.Code), []byte(c.Input().Get("code").String()))
	if err != nil {
		passcode.TryCount += 1
		if passcode.TryCount >= maxPasscodeTries {
			err = deps.Persister.GetPasscodePersister().Delete(*passcode)
			if err != nil {
				return err
			}
			err = c.Stash().Delete("passcode_id")
			if err != nil {
				return err
			}

			return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorPasscodeMaxAttemptsReached)
		}
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorPasscodeInvalid.Wrap(err))
	}

	err = c.Stash().Set("auth_user_id", passcode.UserId)
	if err != nil {
		return fmt.Errorf("failed to set auth_user_id to the stash: %w", err)
	}

	err = c.Stash().Set("email_verified", true) // TODO: maybe change attribute path
	if err != nil {
		return err
	}

	err = deps.Persister.GetPasscodePersisterWithConnection(deps.Tx).Delete(*passcode)
	if err != nil {
		return err
	}

	return c.EndSubFlow()
}
