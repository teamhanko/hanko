package actions

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSendPasscodeCode(passcodeService services.Passcode, httpContext echo.Context) flowpilot.Action {
	return SendPasscodeCode{passcodeService: passcodeService, httpContext: httpContext}
}

type SendPasscodeCode struct {
	httpContext     echo.Context
	passcodeService services.Passcode
}

func (a SendPasscodeCode) GetName() flowpilot.ActionName {
	return common.ActionSendPasscodeCode
}

func (a SendPasscodeCode) GetDescription() string {
	return "Send a passcode code via email."
}

func (a SendPasscodeCode) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a SendPasscodeCode) Execute(c flowpilot.ExecutionContext) error {
	email := c.Stash().Get("email").String()

	acceptLanguageHeader := a.httpContext.Request().Header.Get("Accept-Language")
	passcodeId, err := a.passcodeService.SendLogin(c.GetFlowID(), email, acceptLanguageHeader)
	if err != nil {
		return fmt.Errorf("passcode service failed: %w", err)
	}

	err = c.Stash().Set("passcode_id", passcodeId)
	if err != nil {
		return fmt.Errorf("failed to set passcode_id to stash: %w", err)
	}

	return c.ContinueFlow(common.StateLoginPasscodeConfirmation)
}
