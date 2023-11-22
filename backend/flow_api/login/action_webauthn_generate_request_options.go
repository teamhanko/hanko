package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnGenerateRequestOptions struct {
	shared.Action
}

func (a WebauthnGenerateRequestOptions) GetName() flowpilot.ActionName {
	return ActionWebauthnGenerateRequestOptions
}

func (a WebauthnGenerateRequestOptions) GetDescription() string {
	return "Get webauthn request options in order to sign in with a webauthn credential."
}

func (a WebauthnGenerateRequestOptions) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("webauthn_available").Bool() {
		c.SuspendAction()
	}
}

func (a WebauthnGenerateRequestOptions) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	params := services.GenerateRequestOptionsParams{Tx: deps.Tx}

	sessionDataModel, requestOptions, err := deps.WebauthnService.GenerateRequestOptions(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn request options: %w", err)
	}

	err = c.Stash().Set("webauthn_session_data_id", sessionDataModel.ID)
	if err != nil {
		return fmt.Errorf("failed to stash webauthn_session_data_id: %w", err)
	}

	err = c.Payload().Set("request_options", requestOptions)
	if err != nil {
		return fmt.Errorf("failed to set request_options payload: %w", err)
	}

	return c.ContinueFlow(StateLoginPasskey)
}
