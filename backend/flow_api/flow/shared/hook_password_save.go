package shared

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/flowpilot"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type PasswordSave struct {
	Action
}

func (h PasswordSave) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !c.Stash().Get(StashPathNewPassword).Exists() {
		return nil
	}

	passwordId, _ := uuid.NewV4()
	passwordCredential := models.PasswordCredential{
		ID:       passwordId,
		UserId:   uuid.FromStringOrNil(c.Stash().Get(StashPathUserID).String()),
		Password: c.Stash().Get(StashPathNewPassword).String(),
		TenantID: deps.TenantID,
	}

	err := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).Create(passwordCredential)
	if err != nil {
		return fmt.Errorf("could not create password: %w", err)
	}
	// TODO: add audit log?
	return nil
}
