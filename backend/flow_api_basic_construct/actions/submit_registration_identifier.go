package actions

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"strings"
)

func NewSubmitRegistrationIdentifier(cfg config.Config, persister persistence.Persister, httpContext echo.Context) SubmitRegistrationIdentifier {
	return SubmitRegistrationIdentifier{
		cfg,
		persister,
		httpContext,
	}
}

type SubmitRegistrationIdentifier struct {
	cfg         config.Config
	persister   persistence.Persister
	httpContext echo.Context
}

func (m SubmitRegistrationIdentifier) GetName() flowpilot.ActionName {
	return common.ActionSubmitRegistrationIdentifier
}

func (m SubmitRegistrationIdentifier) GetDescription() string {
	return "Enter at least one identifier to register."
}

func (m SubmitRegistrationIdentifier) Initialize(c flowpilot.InitializationContext) {
	inputs := make([]flowpilot.Input, 0)
	emailInput := flowpilot.EmailInput("email").MaxLength(255).Persist(true).Preserve(true)
	if m.cfg.Identifier.Email.Enabled == "optional" {
		emailInput.Required(false)
		inputs = append(inputs, emailInput)
	} else if m.cfg.Identifier.Email.Enabled == "required" {
		emailInput.Required(true)
		inputs = append(inputs, emailInput)
	}

	usernameInput := flowpilot.StringInput("username").MinLength(2).MaxLength(255).Persist(true).Preserve(true)
	if m.cfg.Identifier.Username.Enabled == "optional" {
		usernameInput.Required(false)
		inputs = append(inputs, usernameInput)
	} else if m.cfg.Identifier.Username.Enabled == "required" {
		usernameInput.Required(true)
		inputs = append(inputs, usernameInput)
	}
	c.AddInputs(inputs...)
}

func (m SubmitRegistrationIdentifier) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	email := c.Input().Get("email").String()
	username := c.Input().Get("username").String()
	//c.Input().SetError("email", flowpilot.ErrorValueMissing)
	//c.Input().SetError("username", flowpilot.ErrorValueMissing)

	for _, char := range username {
		// check that username only contains allowed characters
		if !strings.Contains(m.cfg.Identifier.Username.AllowedCharacters, string(char)) {
			c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username contains invalid characters")))
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if email != "" {
		e, err := m.persister.GetEmailPersister().FindByAddress(email)
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
		}
		// Do not return an error when only identifier is email and email verification is on (account enumeration protection)
		if e != nil && !(m.cfg.Identifier.Username.Enabled == "disabled" && m.cfg.Emails.RequireVerification) {
			c.Input().SetError("email", flowpilot.ErrorValueInvalid.Wrap(errors.New("email already exists"))) // TODO: Maybe create a new error for this
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if username != "" {
		e, err := m.persister.GetUsernamePersister().Find(username)
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical)
		}
		if e != nil {
			c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username already taken"))) // TODO: Maybe create a new error for this
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	err := c.CopyInputValuesToStash("email", "username")
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	if email != "" && m.cfg.Emails.RequireVerification {
		// TODO: send passcode
		return c.ContinueFlow(common.StateEmailVerification)
	} else if m.cfg.Password.Enabled {
		return c.ContinueFlow(common.StatePasswordCreation)
	}

	// TODO: do we need this setting or should we check the config in common.ActionSkip if passkey onboarding is required, although we know it already here
	err = c.Stash().Set("passkey_onboarding_required", true)
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}
	return c.ContinueFlow(common.StateOnboardingCreatePasskey)
}
