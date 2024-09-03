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
	if !c.Stash().Get(shared.StashPathWebauthnAvailable).Bool() {
		return nil
	}

	if !c.Stash().Get(shared.StashPathWebauthnConditionalMediationAvailable).Bool() {
		return nil
	}

	deps := a.GetDeps(c)

	if !deps.Cfg.Passkey.Enabled {
		return nil
	}

	params := services.GenerateRequestOptionsPasskeyParams{Tx: deps.Tx}

	sessionDataModel, requestOptions, err := deps.WebauthnService.GenerateRequestOptionsPasskey(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn request options: %w", err)
	}

	err = c.Stash().Set(shared.StashPathWebauthnSessionDataID, sessionDataModel.ID)
	if err != nil {
		return fmt.Errorf("failed to stash webauthn_session_data_id: %w", err)
	}

	err = c.Payload().Set("request_options", requestOptions)
	if err != nil {
		return fmt.Errorf("failed to set request_options payload: %w", err)
	}

	return nil
}
