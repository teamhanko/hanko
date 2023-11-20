package shared

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SendPasscode struct {
	Action
}

func (h SendPasscode) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !c.Stash().Get("email").Exists() {
		return errors.New("email has not been stashed")
	}

	if !c.Stash().Get("passcode_template").Exists() {
		return errors.New("passcode_template has not been stashed")
	}

	// TODO: rate limit sending emails

	email := c.Stash().Get("email").String()
	template := c.Stash().Get("passcode_template").String()
	acceptLanguageHeader := deps.HttpContext.Request().Header.Get("Accept-Language")

	passcodeId, err := deps.PasscodeService.SendPasscode(c.GetFlowID(), template, email, acceptLanguageHeader)
	if err != nil {
		return fmt.Errorf("passcode service failed: %w", err)
	}

	err = c.Stash().Set("passcode_id", passcodeId)
	if err != nil {
		return fmt.Errorf("failed to set passcode_id to stash: %w", err)
	}

	return nil
}
