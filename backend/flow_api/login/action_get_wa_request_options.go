package login

import (
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type GetWARequestOptions struct {
	shared.Action
}

func (a GetWARequestOptions) GetName() flowpilot.ActionName {
	return shared.ActionGetWARequestOptions
}

func (a GetWARequestOptions) GetDescription() string {
	return "Get request options to use a webauthn credential."
}

func (a GetWARequestOptions) Initialize(c flowpilot.InitializationContext) {
	webAuthnAvailable := c.Stash().Get("webauthn_available").Bool()
	if !webAuthnAvailable {
		c.SuspendAction()
	}
}

func (a GetWARequestOptions) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	options, sessionData, err := deps.Cfg.Webauthn.Handler.BeginDiscoverableLogin(
		webauthnLib.WithUserVerification(protocol.UserVerificationRequirement(deps.Cfg.Webauthn.UserVerification)),
	)
	if err != nil {
		return fmt.Errorf("failed to create webauthn assertion options for discoverable login: %w", err)
	}

	webAuthnSessionDataModel := *intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationAuthentication)
	err = deps.Persister.GetWebauthnSessionDataPersisterWithConnection(deps.Tx).Create(webAuthnSessionDataModel)
	if err != nil {
		return fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	}

	err = c.Stash().Set("webauthn_session_data_id", webAuthnSessionDataModel.ID)
	if err != nil {
		return fmt.Errorf("failed to stash webauthn_session_data_id: %w", err)
	}

	err = c.Payload().Set("request_options", options)
	if err != nil {
		return fmt.Errorf("failed to set request_options payload: %w", err)
	}

	return c.ContinueFlow(StateLoginPasskey)
}
