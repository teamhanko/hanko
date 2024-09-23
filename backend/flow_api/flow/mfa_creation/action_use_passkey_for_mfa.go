package mfa_creation

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type UsePasskeyForMFA struct {
	shared.Action
}

func (a UsePasskeyForMFA) GetName() flowpilot.ActionName {
	return shared.ActionUsePasskeyForMFA
}

func (a UsePasskeyForMFA) GetDescription() string {
	return "Use Passkey for MFA"
}

func (a UsePasskeyForMFA) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get(shared.StashPathUserHasPasskey).Bool() {
		c.SuspendAction()
	}
}

func (a UsePasskeyForMFA) Execute(c flowpilot.ExecutionContext) error {
	if c.IsFlow(shared.FlowLogin) {
		deps := a.GetDeps(c)
		err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Update(models.User{
			ID:               uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String()),
			UsePasskeyForMFA: true,
			UpdatedAt:        time.Now().UTC(),
		})
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	} else if c.IsFlow(shared.FlowRegistration) {
		if err := c.Stash().Set(shared.StashPathUsePasskeyForMFA, true); err != nil {
			return fmt.Errorf("failed to set use_passkey_for_mfa to the stash: %w", err)
		}
	}

	return c.Continue()
}
