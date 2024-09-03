package credential_usage

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnGenerateRequestOptions struct {
	shared.Action
}

func (a WebauthnGenerateRequestOptions) GetName() flowpilot.ActionName {
	return shared.ActionWebauthnGenerateRequestOptions
}

func (a WebauthnGenerateRequestOptions) GetDescription() string {
	return "Get webauthn request options in order to sign in with a webauthn credential."
}

func (a WebauthnGenerateRequestOptions) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !c.Stash().Get(shared.StashPathWebauthnAvailable).Bool() || !deps.Cfg.Passkey.Enabled {
		c.SuspendAction()
	}
}

func (a WebauthnGenerateRequestOptions) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	params := services.GenerateRequestOptionsPasskeyParams{Tx: deps.Tx}

	sessionDataModel, requestOptions, err := deps.WebauthnService.GenerateRequestOptionsPasskey(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn request options: %w", err)
	}

	err = c.Stash().Set(shared.StashPathWebauthnSessionDataID, sessionDataModel.ID)
	if err != nil {
		return fmt.Errorf("failed to stash webauthn_session_data_id: %w", err)
	}

	err = c.Stash().Set(shared.StashPathUserID, sessionDataModel.UserId)
	if err != nil {
		return fmt.Errorf("failed to stash user_id: %w", err)
	}

	err = c.Payload().Set("request_options", requestOptions)
	if err != nil {
		return fmt.Errorf("failed to set request_options payload: %w", err)
	}

	return c.Continue(shared.StateLoginPasskey)
}
