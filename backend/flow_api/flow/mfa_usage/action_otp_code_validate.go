package mfa_usage

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/rate_limiter"
)

type OTPCodeValidate struct {
	shared.Action
}

func (a OTPCodeValidate) GetName() flowpilot.ActionName {
	return shared.ActionOTPCodeValidate
}

func (a OTPCodeValidate) GetDescription() string {
	return "Validates the provided code."
}

func (a OTPCodeValidate) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("otp_code").Required(true))
}

func (a OTPCodeValidate) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	if !c.Stash().Get(shared.StashPathUserID).Exists() {
		return errors.New("user_id does not exist in the stash")
	}

	userID := uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())

	if deps.Cfg.RateLimiter.Enabled {
		rateLimitKey := rate_limiter.CreateRateLimitPasswordKey(deps.HttpContext.RealIP(), userID.String())
		retryAfterSeconds, ok, err := rate_limiter.Limit2(deps.OTPRateLimiter, rateLimitKey)
		if err != nil {
			return fmt.Errorf("rate limiter failed: %w", err)
		}

		if !ok {
			err = c.Payload().Set("retry_after", retryAfterSeconds)
			if err != nil {
				return fmt.Errorf("failed to set a value for retry_after to the payload: %w", err)
			}
			return c.Error(shared.ErrorRateLimitExceeded.Wrap(fmt.Errorf("rate limit exceeded for: %s", rateLimitKey)))
		}
	}

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	code := c.Input().Get("otp_code").String()

	if !totp.Validate(code, userModel.OTPSecret.Secret) {
		return c.Error(shared.ErrorPasscodeInvalid)
	}

	err = c.Stash().Set(shared.StashPathMFAUsageMethod, "totp")
	if err != nil {
		return fmt.Errorf("failed to stash mfa_method: %w", err)
	}

	c.PreventRevert()

	return c.Continue()
}
