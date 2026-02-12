package profile

import (
	"errors"
	"fmt"
	"time"

	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/rate_limiter"
)

type ExchangeToken struct {
	shared.Action
}

func (a ExchangeToken) GetName() flowpilot.ActionName {
	return shared.ActionExchangeToken // TODO: should we rename it?
}

func (a ExchangeToken) GetDescription() string {
	return "Exchange a token for linking the identity to a user account."
}

func (a ExchangeToken) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(
		flowpilot.StringInput("token").Hidden(true).Required(true),
		flowpilot.StringInput("code_verifier").Hidden(true),
	)
}

func (a ExchangeToken) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	deps := a.GetDeps(c)

	if deps.Cfg.RateLimiter.Enabled {
		rateLimitKey := rate_limiter.CreateRateLimitTokenExchangeKey(deps.HttpContext.RealIP())
		retryAfterSeconds, ok, err := rate_limiter.Limit2(deps.TokenExchangeRateLimiter, rateLimitKey)
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

	tokenModel, err := deps.Persister.GetTokenPersisterWithConnection(deps.Tx).GetByValue(c.Input().Get("token").String())
	if err != nil {
		return fmt.Errorf("failed to fetch token from db: %w", err)
	}

	if tokenModel == nil {
		return errors.New("token not found")
	}

	if tokenModel.PKCECodeVerifier != nil && *tokenModel.PKCECodeVerifier != "" && *tokenModel.PKCECodeVerifier != c.Input().Get("code_verifier").String() {
		return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(errors.New("code_verifier does not match")))
	}

	if time.Now().UTC().After(tokenModel.ExpiresAt) {
		return errors.New("token expired")
	}

	user, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	if user.ID != tokenModel.UserID {
		return c.Error(flowpilot.ErrorOperationNotPermitted.Wrap(
			fmt.Errorf("token does not belong to the current user. current user: %s, token user: %s", user.ID, tokenModel.UserID),
		))
	}

	identity, err := deps.Persister.GetIdentityPersisterWithConnection(deps.Tx).GetByID(*tokenModel.IdentityID)
	if err != nil {
		return fmt.Errorf("failed to fetch identity from db: %w", err)
	}

	if identity.UserID != nil {
		return flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("identity is already linked to a user"))
	}

	// link the identity to the user
	identity.UserID = &user.ID

	err = deps.Persister.GetIdentityPersisterWithConnection(deps.Tx).Update(*identity)
	if err != nil {
		return fmt.Errorf("failed to update identity: %w", err)
	}

	err = deps.Persister.GetTokenPersisterWithConnection(deps.Tx).Delete(*tokenModel)
	if err != nil {
		return fmt.Errorf("failed to delete token from db: %w", err)
	}

	c.PreventRevert()

	return c.Continue(shared.StateProfileInit)
}
