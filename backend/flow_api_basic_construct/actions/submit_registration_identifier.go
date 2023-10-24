package actions

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"strings"
)

func NewSubmitRegistrationIdentifier(cfg config.Config, persister persistence.Persister, passcodeService *services.Passcode, httpContext echo.Context) SubmitRegistrationIdentifier {
	return SubmitRegistrationIdentifier{
		cfg,
		persister,
		httpContext,
		passcodeService,
	}
}

// SubmitRegistrationIdentifier takes the identifier which the user entered and checks if they are valid and available according to the configuration
type SubmitRegistrationIdentifier struct {
	cfg             config.Config
	persister       persistence.Persister
	httpContext     echo.Context
	passcodeService *services.Passcode
}

func (m SubmitRegistrationIdentifier) GetName() flowpilot.ActionName {
	return common.ActionSubmitRegistrationIdentifier
}

func (m SubmitRegistrationIdentifier) GetDescription() string {
	return "Enter an identifier to register."
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

	usernameInput := flowpilot.StringInput("username").MinLength(m.cfg.Identifier.Username.MinLength).MaxLength(m.cfg.Identifier.Username.MaxLength).Persist(true).Preserve(true)
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

	for _, char := range username {
		// check that username only contains allowed characters
		if !strings.Contains(m.cfg.Identifier.Username.AllowedCharacters, string(char)) {
			c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username contains invalid characters")))
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if email != "" {
		// Check that email is not already taken
		// this check is non-exhaustive as the email is not blocked here and might be created after the check here and the user creation
		e, err := m.persister.GetEmailPersister().FindByAddress(email)
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
		}
		// Do not return an error when only identifier is email and email verification is on (account enumeration protection)
		if e != nil && !(m.cfg.Identifier.Username.Enabled == "disabled" && m.cfg.Emails.RequireVerification) {
			c.Input().SetError("email", common.ErrorEmailAlreadyExists)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	if username != "" {
		// Check that username is not already taken
		// this check is non-exhaustive as the username is not blocked here and might be created after the check here and the user creation
		e, err := m.persister.GetUsernamePersister().Find(username)
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
		}
		if e != nil {
			c.Input().SetError("username", common.ErrorUsernameAlreadyExists)
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
		}
	}

	err := c.CopyInputValuesToStash("email", "username")
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	// Decide which is the next state according to the config and user input
	if email != "" && m.cfg.Emails.RequireVerification {
		// TODO: rate limit sending emails
		passcodeId, err := m.passcodeService.SendEmailVerification(c.GetFlowID(), email, m.httpContext.Request().Header.Get("Accept-Language"))
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
		}
		err = c.Stash().Set("passcode_id", passcodeId)
		if err != nil {
			return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
		}
		return c.ContinueFlow(common.StateEmailVerification)
	} else if m.cfg.Password.Enabled {
		return c.ContinueFlow(common.StatePasswordCreation)
	} else if !m.cfg.Passcode.Enabled {
		return c.StartSubFlow(common.StateOnboardingCreatePasskey, common.StateSuccess)
	}

	// TODO: store user and create session token // should this case even exist (only works when email (optional/required) is set by the user) ???

	return c.ContinueFlow(common.StateSuccess)
}
