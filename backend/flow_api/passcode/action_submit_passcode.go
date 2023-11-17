package passcode

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var maxPasscodeTries = 3

type SubmitPasscode struct {
	cfg       config.Config
	persister persistence.Persister
}

func (m SubmitPasscode) GetName() flowpilot.ActionName {
	return shared.ActionSubmitPasscode
}

func (m SubmitPasscode) GetDescription() string {
	return "Enter a passcode."
}

func (m SubmitPasscode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("code").Required(true))
}

func (m SubmitPasscode) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	passcodeId, err := uuid.FromString(c.Stash().Get("passcode_id").String())
	if err != nil {
		return err
	}

	passcode, err := m.persister.GetPasscodePersister().Get(passcodeId)
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
			err = m.persister.GetPasscodePersister().Delete(*passcode)
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

	err = c.Stash().Set("email_verified", true) // TODO: maybe change attribute path
	if err != nil {
		return err
	}

	err = m.persister.GetPasscodePersister().Delete(*passcode)
	if err != nil {
		return err
	}

	return c.EndSubFlow()
}
