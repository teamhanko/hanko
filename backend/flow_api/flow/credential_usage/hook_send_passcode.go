package credential_usage

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/rate_limiter"
)

type SendPasscode struct {
	shared.Action
}

func (h SendPasscode) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if c.GetFlowError() != nil {
		return nil
	}

	if !c.Stash().Get(shared.StashPathEmail).Exists() {
		return errors.New("email has not been stashed")
	}

	if !c.Stash().Get(shared.StashPathPasscodeTemplate).Exists() {
		return errors.New("passcode_template has not been stashed")
	}

	if deps.Cfg.RateLimiter.Enabled {
		rateLimitKey := rate_limiter.CreateRateLimitKey(deps.HttpContext.RealIP(), c.Stash().Get(shared.StashPathEmail).String())
		resendAfterSeconds, ok, err := rate_limiter.Limit2(deps.RateLimiter, rateLimitKey)
		if err != nil {
			return fmt.Errorf("rate limiter failed: %w", err)
		}

		if !ok {
			err = c.Payload().Set("resend_after", resendAfterSeconds)
			if err != nil {
				return fmt.Errorf("failed to set a value for resend_after to the payload: %w", err)
			}

			c.SetFlowError(shared.ErrorRateLimitExceeded.Wrap(fmt.Errorf("rate limit exceeded for: %s", rateLimitKey)))
			return nil
		}
	}

	validationParams := services.ValidatePasscodeParams{
		Tx:         deps.Tx,
		PasscodeID: uuid.FromStringOrNil(c.Stash().Get(shared.StashPathPasscodeID).String()),
	}

	passcodeIsValid, err := deps.PasscodeService.ValidatePasscode(validationParams)
	if err != nil {
		return fmt.Errorf("failed to validate existing passcode: %w", err)
	}

	isDifferentEmailAddress := c.Stash().Get(shared.StashPathEmail).String() != c.Stash().Get(shared.StashPathPasscodeEmail).String()

	if !passcodeIsValid || isDifferentEmailAddress {
		sendParams := services.SendPasscodeParams{
			FlowID:       c.GetFlowID(),
			Template:     c.Stash().Get(shared.StashPathPasscodeTemplate).String(),
			EmailAddress: c.Stash().Get(shared.StashPathEmail).String(),
			Language:     deps.HttpContext.Request().Header.Get("Accept-Language"),
		}

		passcodeID, err := deps.PasscodeService.SendPasscode(sendParams)
		if err != nil {
			return fmt.Errorf("passcode service failed: %w", err)
		}

		err = c.Stash().Set(shared.StashPathPasscodeID, passcodeID)
		if err != nil {
			return fmt.Errorf("failed to set passcode_id to stash: %w", err)
		}

		err = c.Stash().Set(shared.StashPathPasscodeEmail, c.Stash().Get(shared.StashPathEmail).String())
		if err != nil {
			return fmt.Errorf("failed to set passcode_email to stash: %w", err)
		}
	}

	return nil
}
