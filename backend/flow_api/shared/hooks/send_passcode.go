package hooks

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SendPasscode struct {
	httpContext     echo.Context
	passcodeService services.Passcode
}

func (a SendPasscode) Execute(c flowpilot.HookExecutionContext) error {
	if !c.Stash().Get("email").Exists() {
		return errors.New("email has not been stashed")
	}

	if !c.Stash().Get("passcode_template").Exists() {
		return errors.New("passcode_template has not been stashed")
	}

	// TODO: rate limit sending emails

	email := c.Stash().Get("email").String()
	template := c.Stash().Get("passcode_template").String()
	acceptLanguageHeader := a.httpContext.Request().Header.Get("Accept-Language")

	passcodeId, err := a.passcodeService.SendPasscode(c.GetFlowID(), template, email, acceptLanguageHeader)
	if err != nil {
		return fmt.Errorf("passcode service failed: %w", err)
	}

	err = c.Stash().Set("passcode_id", passcodeId)
	if err != nil {
		return fmt.Errorf("failed to set passcode_id to stash: %w", err)
	}

	return nil
}
