package hooks

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSendPasscode(passcodeService services.Passcode, httpContext echo.Context) flowpilot.HookAction {
	return SendPasscode{httpContext: httpContext, passcodeService: passcodeService}
}

type SendPasscode struct {
	httpContext     echo.Context
	passcodeService services.Passcode
}

func (a SendPasscode) Execute(c flowpilot.HookExecutionContext) error {
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

	return nil
}
