package shared

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type IssueSession struct {
	Action
}

func (h IssueSession) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	var userId uuid.UUID
	var err error
	if c.Stash().Get(StashPathUserID).Exists() {
		userId, err = uuid.FromString(c.Stash().Get(StashPathUserID).String())
		if err != nil {
			return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
		}
	} else {
		return errors.New("user_id not found in stash")
	}

	emails, err := deps.Persister.GetEmailPersisterWithConnection(deps.Tx).FindByUserId(userId)
	if err != nil {
		return fmt.Errorf("failed to fetch emails from db: %w", err)
	}

	var emailDTO *dto.EmailJwt

	if email := emails.GetPrimary(); email != nil {
		emailDTO = dto.JwtFromEmailModel(email)
	}

	sessionToken, err := deps.SessionManager.GenerateJWT(userId, emailDTO)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	cookie, err := deps.SessionManager.GenerateCookie(sessionToken)
	if err != nil {
		return fmt.Errorf("failed to generate auth cookie, %w", err)
	}

	deps.HttpContext.Response().Header().Set("X-Session-Lifetime", fmt.Sprintf("%d", cookie.MaxAge))

	if deps.Cfg.Session.EnableAuthTokenHeader {
		deps.HttpContext.Response().Header().Set("X-Auth-Token", sessionToken)
	} else {
		deps.HttpContext.SetCookie(cookie)
	}

	// Audit log logins only, because user creation on registration implies that the user is logged
	// in after a registration. Only login actions should set the "login_method" stash entry.
	if c.Stash().Get(StashPathLoginMethod).Exists() {
		err = deps.AuditLogger.CreateWithConnection(
			deps.Tx,
			deps.HttpContext,
			models.AuditLogLoginSuccess,
			&models.User{ID: userId},
			err,
			auditlog.Detail("login_method", c.Stash().Get(StashPathLoginMethod).String()),
			auditlog.Detail("flow_id", c.GetFlowID()))

		if err != nil {
			return fmt.Errorf("could not create audit log: %w", err)
		}
	}

	return nil
}
