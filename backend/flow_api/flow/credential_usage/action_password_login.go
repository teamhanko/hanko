package credential_usage

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/rate_limiter"
)

type PasswordLogin struct {
	shared.Action
}

func (a PasswordLogin) GetName() flowpilot.ActionName {
	return shared.ActionPasswordLogin
}

func (a PasswordLogin) GetDescription() string {
	return "Login with a password."
}

func (a PasswordLogin) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.PasswordInput("password").Required(true))

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	}
}

func (a PasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	if deps.Cfg.RateLimiter.Enabled {
		rateLimitKey := rate_limiter.CreateRateLimitPasswordKey(deps.HttpContext.RealIP(), c.Stash().Get(shared.StashPathUserIdentification).String())
		retryAfterSeconds, ok, err := rate_limiter.Limit2(deps.PasswordRateLimiter, rateLimitKey)
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

	var userID uuid.UUID

	if c.Stash().Get(shared.StashPathEmail).Exists() {
		emailModel, err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).FindByAddress(c.Stash().Get(shared.StashPathEmail).String())
		if err != nil {
			return fmt.Errorf("failed to find user by email: %w", err)
		}

		if emailModel == nil {
			return a.wrongCredentialsError(c)
		}

		userID = *emailModel.UserID
	} else if c.Stash().Get(shared.StashPathUsername).Exists() {
		username := c.Stash().Get(shared.StashPathUsername).String()
		userModel, err := deps.Persister.GetUserPersister().GetByUsername(username)
		if err != nil {
			return fmt.Errorf("failed to find user via username: %w", err)
		}

		if userModel == nil {
			return a.wrongCredentialsError(c)
		}

		userID = userModel.ID
	} else {
		return a.wrongCredentialsError(c)
	}

	err := deps.PasswordService.VerifyPassword(deps.Tx, userID, c.Input().Get("password").String())
	if err != nil {
		if errors.Is(err, services.ErrorPasswordInvalid) {
			err = deps.AuditLogger.CreateWithConnection(
				deps.Tx,
				deps.HttpContext,
				models.AuditLogLoginFailure,
				&models.User{ID: userID},
				err,
				auditlog.Detail("login_method", "password"),
				auditlog.Detail("flow_id", c.GetFlowID()))
			if err != nil {
				return fmt.Errorf("could not create audit log: %w", err)
			}

			return a.wrongCredentialsError(c)
		}

		return fmt.Errorf("failed to verify password: %w", err)
	}

	// Set only for audit logging purposes.
	err = c.Stash().Set(shared.StashPathLoginMethod, "password")
	if err != nil {
		return fmt.Errorf("failed to set login_method to the stash: %w", err)
	}

	c.PreventRevert()

	return c.Continue()
}

func (a PasswordLogin) wrongCredentialsError(c flowpilot.ExecutionContext) error {
	c.Input().SetError("password", flowpilot.ErrorValueInvalid)
	return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(errors.New("wrong credentials")))
}
