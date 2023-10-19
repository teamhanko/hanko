package actions

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var maxPasscodeTries = 3

func NewSubmitPasscode(cfg config.Config, persister persistence.Persister, userService services.User, sessionManager session.Manager, httpContext echo.Context) SubmitPasscode {
	return SubmitPasscode{
		cfg,
		persister,
		userService,
		sessionManager,
		httpContext,
	}
}

type SubmitPasscode struct {
	cfg            config.Config
	persister      persistence.Persister
	userService    services.User
	sessionManager session.Manager
	httpContext    echo.Context
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
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	passcode, err := m.persister.GetPasscodePersister().Get(passcodeId)
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
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
				return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
			}
			err = c.Stash().Delete("passcode_id")
			if err != nil {
				return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
			}

			return c.ContinueFlowWithError(c.GetCurrentState(), common.ErrorPasscodeMaxAttemptsReached)
		}
		return c.ContinueFlowWithError(c.GetCurrentState(), common.ErrorPasscodeInvalid.Wrap(err))
	}

	err = c.Stash().Set("email_verified", true) // TODO: maybe change attribute path
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	err = m.persister.GetPasscodePersister().Delete(*passcode)
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	// TODO: This the current routing is only for the registration flow, when this action is/will be used in the login flow on other states, then the routing needs to be changed accordingly
	// Decide which is the next state according to the config and user input
	if m.cfg.Password.Enabled {
		return c.ContinueFlow(common.StatePasswordCreation)
	} else /*if m.cfg.SecondFactor.Enabled != "disabled" {
		var capabilities capabilities
		err = json.Unmarshal([]byte(c.Stash().Get("capabilities").String()), &capabilities)
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
		}

		nextStates := []flowpilot.StateName{common.StateSuccess}
		if capabilities.Webauthn.Available && m.cfg.Passkey.Onboarding.Enabled {
			nextStates = append(nextStates, common.StateOnboardingCreatePasskey)
		}
		if capabilities.Webauthn.Available && slices.Contains(m.cfg.SecondFactor.Methods, "security_key") {
			return c.StartSubFlow(common.StateCreate2FASecurityKey, nextStates...)
		} else if slices.Contains(m.cfg.SecondFactor.Methods, "totp") {
			// TODO: This does not work, as a subflow only has ONE init state, but here we need two
			return c.StartSubFlow(common.StateCreate2FATOTP, nextStates...)
		} else {
			// This case should never occur. The config validation should catch this case.
			// No 2FA method is configured. At least on method must be configured when 2FA is enabled (optional/required).
			return c.ContinueFlowWithError(c.GetErrorState(), common.ErrorConfigurationError)
		}
	} else*/if !m.cfg.Passcode.Enabled || m.cfg.Passkey.Onboarding.Enabled {
		return c.StartSubFlow(common.StateOnboardingCreatePasskey, common.StateSuccess)
	}

	// store user in the DB
	userId, err := uuid.NewV4()
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}
	if c.Stash().Get("user_id").Exists() {
		userId, err = uuid.FromString(c.Stash().Get("user_id").String())
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
		}
	}
	err = m.userService.CreateUser(
		userId,
		c.Stash().Get("email").String(),
		c.Stash().Get("email_verified").Bool(),
		c.Stash().Get("username").String(),
		nil,
		c.Stash().Get("new_password").String(),
	)
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	sessionToken, err := m.sessionManager.GenerateJWT(userId)
	if err != nil {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorTechnical.Wrap(err))
	}
	cookie, err := m.sessionManager.GenerateCookie(sessionToken)
	if err != nil {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	m.httpContext.SetCookie(cookie)

	return c.ContinueFlow(common.StateSuccess)
}
