package mfa_usage

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v3/flow_api/services"
	"github.com/teamhanko/hanko/backend/v3/flowpilot"
	"github.com/teamhanko/hanko/backend/v3/rate_limiter"
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
		rateLimitKey := rate_limiter.CreateRateLimitOTPKey(deps.HttpContext.RealIP(), userID.String())
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

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userID, deps.TenantID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	code := c.Input().Get("otp_code").String()

	// The secret is locked until deps.Tx ends, so a concurrent flow of the same user waits here and then
	// validates against the step this flow consumed. The preloaded userModel.OTPSecret may be stale by then.
	otpSecretPersister := deps.Persister.GetOTPSecretPersisterWithConnection(deps.Tx)
	otpSecret, err := otpSecretPersister.GetForUpdate(userModel.OTPSecret.ID, deps.TenantID)
	if err != nil {
		return fmt.Errorf("failed to lock otp secret: %w", err)
	}

	matchedStep, err := deps.TOTPService.ValidateCode(code, otpSecret, time.Now().UTC())
	if err != nil {
		if errors.Is(err, services.ErrTOTPCodeAlreadyUsed) {
			deps.HttpContext.Logger().Warn(fmt.Errorf("totp replay attempt for user %s: %w", userID, err))
			c.Input().SetError("otp_code", shared.ErrorOTPCodeAlreadyUsed)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}
		if errors.Is(err, services.ErrTOTPCodeInvalid) {
			c.Input().SetError("otp_code", shared.ErrorOTPCodeInvalid)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}
		return fmt.Errorf("totp validation error: %w", err)
	}

	// Persist the consumed step atomically within deps.Tx (same transaction as session creation).
	otpSecret.LastValidatedStep = nulls.NewInt64(matchedStep)
	if err := otpSecretPersister.Update(otpSecret); err != nil {
		return fmt.Errorf("failed to update otp secret last validated step: %w", err)
	}

	err = c.Stash().Set(shared.StashPathMFAUsageMethod, "totp")
	if err != nil {
		return fmt.Errorf("failed to stash mfa_method: %w", err)
	}

	c.PreventRevert()

	return c.Continue()
}
