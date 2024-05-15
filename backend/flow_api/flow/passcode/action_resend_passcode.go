package passcode

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/rate_limiter"
)

type ReSendPasscode struct {
	shared.Action
}

func (a ReSendPasscode) GetName() flowpilot.ActionName {
	return constants.ActionResendPasscode
}

func (a ReSendPasscode) GetDescription() string {
	return "Send the passcode email again."
}

func (a ReSendPasscode) Initialize(_ flowpilot.InitializationContext) {}

func (a ReSendPasscode) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if !c.Stash().Get("email").Exists() {
		return errors.New("email has not been stashed")
	}

	if !c.Stash().Get("passcode_template").Exists() {
		return errors.New("passcode_template has not been stashed")
	}

	if deps.Cfg.RateLimiter.Enabled {
		rateLimitKey := rate_limiter.CreateRateLimitKey(deps.HttpContext.RealIP(), c.Stash().Get("email").String())
		resendAfterSeconds, ok, err := rate_limiter.Limit2(deps.RateLimiter, rateLimitKey)
		if err != nil {
			return fmt.Errorf("rate limiter failed: %w", err)
		}

		if !ok {
			err = c.Payload().Set("resend_after", resendAfterSeconds)
			if err != nil {
				return fmt.Errorf("failed to set a value for resend_after to the payload: %w", err)
			}
			return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorRateLimitExceeded.Wrap(fmt.Errorf("rate limit exceeded for: %s", rateLimitKey)))
		}
	}

	sendParams := services.SendPasscodeParams{
		FlowID:       c.GetFlowID(),
		Template:     c.Stash().Get("passcode_template").String(),
		EmailAddress: c.Stash().Get("email").String(),
		Language:     deps.HttpContext.Request().Header.Get("Accept-Language"),
	}
	passcodeID, err := deps.PasscodeService.SendPasscode(sendParams)
	if err != nil {
		return fmt.Errorf("passcode service failed: %w", err)
	}

	err = c.Stash().Set("passcode_id", passcodeID)
	if err != nil {
		return fmt.Errorf("failed to set passcode_id to stash: %w", err)
	}

	return c.ContinueFlow(c.GetCurrentState())
}

func (a ReSendPasscode) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
