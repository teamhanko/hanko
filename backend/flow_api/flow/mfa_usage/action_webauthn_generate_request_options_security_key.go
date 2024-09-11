package mfa_usage

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnGenerateRequestOptionsSecurityKey struct {
	shared.Action
}

func (a WebauthnGenerateRequestOptionsSecurityKey) GetName() flowpilot.ActionName {
	return shared.ActionWebauthnGenerateRequestOptions
}

func (a WebauthnGenerateRequestOptionsSecurityKey) GetDescription() string {
	return "Get webauthn request options in order to sign in with a webauthn credential."
}

func (a WebauthnGenerateRequestOptionsSecurityKey) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get(shared.StashPathWebauthnAvailable).Bool() {
		c.SuspendAction()
	}
}

func (a WebauthnGenerateRequestOptionsSecurityKey) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	params := services.GenerateRequestOptionsSecurityKeyParams{
		Tx:     deps.Tx,
		UserID: uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String()),
	}

	sessionDataModel, requestOptions, err := deps.WebauthnService.GenerateRequestOptionsSecurityKey(params)
	if err != nil {
		return fmt.Errorf("failed to generate webauthn request options: %w", err)
	}

	err = c.Stash().Set(shared.StashPathWebauthnSessionDataID, sessionDataModel.ID)
	if err != nil {
		return fmt.Errorf("failed to stash webauthn_session_data_id: %w", err)
	}

	err = c.Stash().Set(shared.StashPathMFAMethod, "security_key")
	if err != nil {
		return fmt.Errorf("failed to stash mfa_method: %w", err)
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
