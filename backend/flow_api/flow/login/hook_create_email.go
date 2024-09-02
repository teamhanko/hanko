package login

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
)

type CreateEmail struct {
	shared.Action
}

func (h CreateEmail) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !c.Stash().Get(shared.StashPathEmail).Exists() || (deps.Cfg.Email.RequireVerification && !c.Stash().Get(shared.StashPathEmailVerified).Bool()) {
		return nil
	}

	if !c.Stash().Get(shared.StashPathLoginOnboardingCreateEmail).Bool() {
		return nil
	}

	if err := c.Stash().Delete(shared.StashPathLoginOnboardingCreateEmail); err != nil {
		return fmt.Errorf("failed to delete login_onboarding_create_email from the stash: %w", err)
	}

	userID := uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())
	emailModel := models.NewEmail(&userID, c.Stash().Get(shared.StashPathEmail).String())

	err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Create(*emailModel)
	if err != nil {
		return fmt.Errorf("failed to create a new email: %w", err)
	}

	primaryEmail := models.NewPrimaryEmail(emailModel.ID, userID)
	err = deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Create(*primaryEmail)
	if err != nil {
		return fmt.Errorf("failed to create a new primary email: %w", err)
	}

	utils.NotifyUserChange(deps.HttpContext, deps.Tx, deps.Persister, events.UserEmailCreate, userID)

	return nil
}
