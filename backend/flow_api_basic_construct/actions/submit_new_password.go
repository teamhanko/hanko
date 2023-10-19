package actions

import (
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/session"
	"golang.org/x/crypto/bcrypt"
	"unicode/utf8"
)

func NewSubmitNewPassword(cfg config.Config, userService services.User, sessionManager session.Manager, httpContext echo.Context) SubmitNewPassword {
	return SubmitNewPassword{
		cfg,
		userService,
		sessionManager,
		httpContext,
	}
}

type SubmitNewPassword struct {
	cfg            config.Config
	userService    services.User
	sessionManager session.Manager
	httpContext    echo.Context
}

func (m SubmitNewPassword) GetName() flowpilot.ActionName {
	return common.ActionSubmitPassword
}

func (m SubmitNewPassword) GetDescription() string {
	return "Submit a new password."
}

func (m SubmitNewPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("new_password").Required(true).MinLength(m.cfg.Password.MinPasswordLength))
}

func (m SubmitNewPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	newPassword := c.Input().Get("new_password").String()
	newPasswordBytes := []byte(newPassword)
	if utf8.RuneCountInString(newPassword) < m.cfg.Password.MinPasswordLength {
		c.Input().SetError("new_password", flowpilot.ErrorValueInvalid)
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if len(newPasswordBytes) > 72 {
		c.Input().SetError("new_password", flowpilot.ErrorValueInvalid)
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(c.Input().Get("new_password").String()), 12)
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}
	err = c.Stash().Set("new_password", string(hashedPassword))

	// Decide which is the next state according to the config and user input
	if m.cfg.Passkey.Onboarding.Enabled {
		return c.StartSubFlow(common.StateOnboardingCreatePasskey, common.StateSuccess)
	}
	// TODO: 2FA routing

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
