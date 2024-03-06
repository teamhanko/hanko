package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnGenerateRequestOptionsForConditionalUi struct {
	shared.Action
}

func (a WebauthnGenerateRequestOptionsForConditionalUi) Execute(c flowpilot.HookExecutionContext) error {
	if !c.Stash().Get("webauthn_available").Bool() {
		return nil
	}

	if !c.Stash().Get("webauthn_conditional_mediation_available").Bool() {
		return nil
	}

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

	err = c.Stash().Set("user_id", sessionDataModel.UserId)
	if err != nil {
		return fmt.Errorf("failed to stash user_id: %w", err)
	}

	err = c.Payload().Set("request_options", requestOptions)
	if err != nil {
		return fmt.Errorf("failed to set request_options payload: %w", err)
	}

	return nil
}
