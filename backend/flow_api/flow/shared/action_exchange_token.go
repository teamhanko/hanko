package shared

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot"
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
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	deps := a.GetDeps(c)

	tokenModel, terr := deps.Persister.GetTokenPersisterWithConnection(deps.Tx).GetByValue(c.Input().Get("token").String())
	if terr != nil {
		return fmt.Errorf("failed to fetch token from db: %w", terr)
	}

	if tokenModel == nil {
		return errors.New("token not found")
	}

	if time.Now().UTC().After(tokenModel.ExpiresAt) {
		return errors.New("token expired")
	}

	terr = deps.Persister.GetTokenPersisterWithConnection(deps.Tx).Delete(*tokenModel)
	if terr != nil {
		return fmt.Errorf("failed to delete token from db: %w", terr)
	}

	// Set because the thirdparty/callback endpoint already creates a user.
	if err := c.Stash().Set("skip_user_creation", true); err != nil {
		return fmt.Errorf("failed to set skip_user_creation to stash: %w", err)
	}

	// Set so the issue_session hook knows who to create the session for.
	if err := c.Stash().Set("user_id", tokenModel.UserID.String()); err != nil {
		return fmt.Errorf("failed to set user_id to stash: %w", err)
	}

	return c.ContinueFlow(StateSuccess)
}

func (a ExchangeToken) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
