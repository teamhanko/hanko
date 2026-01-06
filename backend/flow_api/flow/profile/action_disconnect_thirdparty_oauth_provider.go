package profile

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type DisconnectThirdpartyOauthProvider struct {
	shared.Action
}

func (a DisconnectThirdpartyOauthProvider) GetName() flowpilot.ActionName {
	return shared.ActionDisconnectThirdpartyOauthProvider
}

func (a DisconnectThirdpartyOauthProvider) GetDescription() string {
	return "Disconnect a third party provider via OAuth."
}

func (a DisconnectThirdpartyOauthProvider) Initialize(c flowpilot.InitializationContext) {
	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if len(userModel.Identities) == 0 {
		c.SuspendAction()
		return
	}

	input := flowpilot.StringInput("identity_id").Required(true)

	for _, identity := range userModel.Identities {
		input.AllowedValue(identity.ProviderID, identity.ID)
	}
}

func (a DisconnectThirdpartyOauthProvider) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	identityIDStr := c.Input().Get("identity_id").String()
	identityID, err := uuid.FromString(identityIDStr)
	if err != nil {
		return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(err))
	}

	if !slices.ContainsFunc(userModel.Identities, func(identity models.Identity) bool {
		return identity.ID == identityID
	}) {
		return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(fmt.Errorf("identity not found")))
	}

	identityPersister := deps.Persister.GetIdentityPersisterWithConnection(deps.Tx)

	identity, err := identityPersister.GetByID(identityID)
	if err != nil {
		return fmt.Errorf("failed to fetch identity from db: %w", err)
	}
	if identity == nil {
		return errors.New("identity not found")
	}

	err = identityPersister.Delete(*identity)
	if err != nil {
		return fmt.Errorf("failed to delete identity from db: %w", err)
	}

	return c.Continue(shared.StateProfileInit)
}
