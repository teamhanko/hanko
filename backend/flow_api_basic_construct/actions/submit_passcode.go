package actions

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var maxPasscodeTries = 3

func NewSubmitPasscode(cfg config.Config, persister persistence.Persister) SubmitPasscode {
	return SubmitPasscode{
		cfg,
		persister,
	}
}

type SubmitPasscode struct {
	cfg       config.Config
	persister persistence.Persister
}

func (m SubmitPasscode) GetName() flowpilot.ActionName {
	return common.ActionSubmitPasscode
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

			return c.ContinueFlowWithError(c.GetCurrentState(), common.ErrorPasscodeMaxAttemptsReached)
		}
		return c.ContinueFlowWithError(c.GetCurrentState(), common.ErrorPasscodeInvalid.Wrap(err))
	}

	err = c.Stash().Set("email_verified", true) // TODO: maybe change attribute path
	if err != nil {
		return err
	}

	err = m.persister.GetPasscodePersister().Delete(*passcode)
	if err != nil {
		return err
	}

	switch c.GetCurrentState() {
	case common.StateRegistrationPasscodeConfirmation:
		// TODO: This the current routing is only for the registration flow, when this action is/will be used in the login flow on other states, then the routing needs to be changed accordingly
		// Decide which is the next state according to the config and user input
		if m.cfg.Password.Enabled {
			return c.ContinueFlow(common.StatePasswordCreation)
		} else if !m.cfg.Passcode.Enabled || (m.cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool()) {
			return c.StartSubFlow(common.StateOnboardingCreatePasskey, common.StateSuccess)
		}
	case common.StateLoginPasscodeConfirmation:
		if m.cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
			return c.StartSubFlow(common.StateOnboardingCreatePasskey, common.StateSuccess)
		}

		return c.ContinueFlow(common.StateSuccess)
	case common.StateLoginPasscodeConfirmationRecovery:
		return c.ContinueFlow(common.StateLoginPasswordRecovery)
	}

	return flowpilot.ErrorFlowDiscontinuity
}
