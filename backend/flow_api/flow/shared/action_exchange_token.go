package shared

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type ExchangeToken struct {
	Action
}

func (a ExchangeToken) GetName() flowpilot.ActionName {
	return ActionExchangeToken
}

func (a ExchangeToken) GetDescription() string {
	return "Exchange a one time token."
}

func (a ExchangeToken) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("token").Hidden(true).Required(true))
}

func (a ExchangeToken) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	deps := a.GetDeps(c)

	tokenModel, err := deps.Persister.GetTokenPersisterWithConnection(deps.Tx).GetByValue(c.Input().Get("token").String())
	if err != nil {
		return fmt.Errorf("failed to fetch token from db: %w", err)
	}

	if tokenModel == nil {
		return errors.New("token not found")
	}

	if time.Now().UTC().After(tokenModel.ExpiresAt) {
		return errors.New("token expired")
	}

	identity, err := deps.Persister.GetIdentityPersisterWithConnection(deps.Tx).GetByID(*tokenModel.IdentityID)
	if err != nil {
		return fmt.Errorf("failed to fetch identity from db: %w", err)
	}

	// Set so the issue_session hook knows who to create the session for.
	if err := c.Stash().Set(StashPathUserID, tokenModel.UserID.String()); err != nil {
		return fmt.Errorf("failed to set user_id to stash: %w", err)
	}

	// Set because the thirdparty/callback endpoint already creates a user.
	if err := c.Stash().Set(StashPathSkipUserCreation, true); err != nil {
		return fmt.Errorf("failed to set skip_user_creation to stash: %w", err)
	}

	err = deps.Persister.GetTokenPersisterWithConnection(deps.Tx).Delete(*tokenModel)
	if err != nil {
		return fmt.Errorf("failed to delete token from db: %w", err)
	}

	onboardingStates, err := a.determineOnboardingStates(c, identity, tokenModel.UserCreated)
	if err != nil {
		return fmt.Errorf("failed to determine onboarding stattes: %w", err)
	}

	return c.Continue(onboardingStates...)
}

func (a ExchangeToken) determineOnboardingStates(c flowpilot.ExecutionContext, identity *models.Identity, userCreated bool) ([]flowpilot.StateName, error) {
	deps := a.GetDeps(c)
	result := make([]flowpilot.StateName, 0)

	if deps.Cfg.Email.RequireVerification && identity.Email != nil && !identity.Email.Verified {
		if err := c.Stash().Set(StashPathEmail, identity.Email.Address); err != nil {
			return nil, fmt.Errorf("failed to stash email: %w", err)
		}

		if err := c.Stash().Set(StashPathPasscodeTemplate, "email_verification"); err != nil {
			return nil, fmt.Errorf("failed to stash passcode_template: %w", err)
		}

		result = append(result, StatePasscodeConfirmation)
	}

	if deps.Cfg.Username.Enabled && len(identity.Email.User.Username.String) == 0 {
		if (!userCreated && deps.Cfg.Username.AcquireOnLogin) ||
			(userCreated && deps.Cfg.Username.AcquireOnRegistration) {
			result = append(result, StateOnboardingUsername)
		}
	}

	return append(result, StateSuccess), nil
}
