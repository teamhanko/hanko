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

	passwordId, _ := uuid.NewV4()
	passwordCredential := models.PasswordCredential{
		ID:       passwordId,
		UserId:   uuid.FromStringOrNil(c.Stash().Get("user_id").String()),
		Password: c.Stash().Get("new_password").String(),
	}
	err := deps.Persister.GetPasswordCredentialPersister().Create(passwordCredential)
	if err != nil {
		return fmt.Errorf("could not create passcode: %w", err)
	}
	// TODO: add audit log?
	return nil
}
