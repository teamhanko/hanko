package shared

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailPersistVerifiedStatus struct {
	Action
}

func (h EmailPersistVerifiedStatus) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !c.Stash().Get("email_verified").Bool() {
		return errors.New("no email has been confirmed as verified")
	}

	if !c.Stash().Get("email").Exists() {
		return errors.New("verified email not set on the stash")
	}

	if !c.Stash().Get("user_id").Exists() {
		return errors.New("user_id not set on the stash")
	}

	userId, err := uuid.FromString(c.Stash().Get("user_id").String())
	if err != nil {
		return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
	}

	emailAddressToVerify := c.Stash().Get("email").String()

	emailAddressToVerifyModel, err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).FindByAddress(emailAddressToVerify)
	if err != nil {
		return fmt.Errorf("could not fetch email: %w", err)
	}

	var emailCreated bool
	if emailAddressToVerifyModel == nil {
		newEmailModel := models.NewEmail(&userId, emailAddressToVerify)
		newEmailModel.Verified = true

		err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Create(*newEmailModel)
		if err != nil {
			return fmt.Errorf("could not save email: %w", err)
		}

		emailModels, err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).FindByUserId(*newEmailModel.UserID)
		if err != nil {
			return fmt.Errorf("could fetch emails: %w", err)
		}

		if len(emailModels) == 1 && emailModels[0].ID.String() == newEmailModel.ID.String() {
			// The user has only one 1 email and it is the email we just added. It makes sense then,
			// to automatically set this as the primary email.
			primaryEmailModel := models.NewPrimaryEmail(newEmailModel.ID, userId)
			err = deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Create(*primaryEmailModel)
			if err != nil {
				return fmt.Errorf("could not save primary email: %w", err)
			}
		}

		emailCreated = true
	} else if !emailAddressToVerifyModel.Verified {
		emailAddressToVerifyModel.Verified = true
		err = deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Update(*emailAddressToVerifyModel)
		if err != nil {
			return fmt.Errorf("could not update email: %w", err)
		}
	}

	// Audit log verification only if this is not a login via passcode because it implies verification.
	// Only login actions should set the "login_method" stash entry.
	if c.Stash().Get("login_method").String() != "passcode" {
		err = deps.AuditLogger.CreateWithConnection(
			deps.Tx,
			deps.HttpContext,
			models.AuditLogEmailVerified,
			&models.User{ID: userId},
			nil,
			auditlog.Detail("email", emailAddressToVerify),
			auditlog.Detail("flow_id", c.GetFlowID()))

		if err != nil {
			return fmt.Errorf("could not create audit log: %w", err)
		}
	}

	if emailCreated {
		err = deps.AuditLogger.CreateWithConnection(
			deps.Tx,
			deps.HttpContext,
			models.AuditLogEmailCreated,
			&models.User{ID: userId},
			nil,
			auditlog.Detail("email", emailAddressToVerify),
			auditlog.Detail("flow_id", c.GetFlowID()))

		if err != nil {
			return fmt.Errorf("could not create audit log: %w", err)
		}
	}

	return nil
}
