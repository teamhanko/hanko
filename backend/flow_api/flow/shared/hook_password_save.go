package shared

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
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
	}

	err := deps.Persister.GetPasswordCredentialPersister().Create(passwordCredential)
	if err != nil {
		return fmt.Errorf("could not create password: %w", err)
	}
	// TODO: add audit log?
	return nil
}
