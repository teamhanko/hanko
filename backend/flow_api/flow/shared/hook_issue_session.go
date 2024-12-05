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

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userId)
	if err != nil {
		return fmt.Errorf("failed to fetch user from db: %w", err)
	}

	var emailDTO *dto.EmailJwt
	if email := userModel.Emails.GetPrimary(); email != nil {
		emailDTO = dto.JwtFromEmailModel(email)
	}

	signedSessionToken, rawToken, err := deps.SessionManager.GenerateJWT(userId, emailDTO)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	activeSessions, err := deps.Persister.GetSessionPersisterWithConnection(deps.Tx).ListActive(userId)
	if err != nil {
		return fmt.Errorf("failed to list active sessions: %w", err)
	}

	if deps.Cfg.Session.ServerSide.Enabled {
		// remove all server side sessions that exceed the limit
		if len(activeSessions) >= deps.Cfg.Session.ServerSide.Limit {
			for i := deps.Cfg.Session.ServerSide.Limit - 1; i < len(activeSessions); i++ {
				err = deps.Persister.GetSessionPersisterWithConnection(deps.Tx).Delete(activeSessions[i])
				if err != nil {
					return fmt.Errorf("failed to remove latest session: %w", err)
				}
			}
		}

		sessionID, _ := rawToken.Get("session_id")

		expirationTime := rawToken.Expiration()
		sessionModel := models.Session{
			ID:        uuid.FromStringOrNil(sessionID.(string)),
			UserID:    userId,
			UserAgent: deps.HttpContext.Request().UserAgent(),
			IpAddress: deps.HttpContext.RealIP(),
			CreatedAt: rawToken.IssuedAt(),
			UpdatedAt: rawToken.IssuedAt(),
			ExpiresAt: &expirationTime,
			LastUsed:  rawToken.IssuedAt(),
		}

		err = deps.Persister.GetSessionPersisterWithConnection(deps.Tx).Create(sessionModel)
		if err != nil {
			return fmt.Errorf("failed to store session: %w", err)
		}
	}

	cookie, err := deps.SessionManager.GenerateCookie(signedSessionToken)
	if err != nil {
		return fmt.Errorf("failed to generate auth cookie, %w", err)
	}

	deps.HttpContext.Response().Header().Set("X-Session-Lifetime", fmt.Sprintf("%d", cookie.MaxAge))

	if deps.Cfg.Session.EnableAuthTokenHeader {
		deps.HttpContext.Response().Header().Set("X-Auth-Token", signedSessionToken)
	} else {
		deps.HttpContext.SetCookie(cookie)
	}

	loginMethod := c.Stash().Get(StashPathLoginMethod)
	mfaMethod := c.Stash().Get(StashPathMFAUsageMethod)
	thirdPartyProvider := c.Stash().Get(StashPathThirdPartyProvider)

	// Audit log logins only, because user creation on registration implies that the user is logged
	// in after a registration. Only login actions should set the "login_method" stash entry.
	if loginMethod.Exists() {
		auditLogDetails := []auditlog.DetailOption{
			auditlog.Detail("login_method", loginMethod.String()),
			auditlog.Detail("flow_id", c.GetFlowID()),
		}

		if mfaMethod.Exists() {
			auditLogDetails = append(
				auditLogDetails,
				auditlog.Detail("mfa_method", mfaMethod.String()),
			)
		}

		err = deps.AuditLogger.CreateWithConnection(
			deps.Tx,
			deps.HttpContext,
			models.AuditLogLoginSuccess,
			&models.User{ID: userId},
			err,
			auditLogDetails...)

		if err != nil {
			return fmt.Errorf("could not create audit log: %w", err)
		}
	}

	if loginMethod.Exists() {
		if err := c.Payload().Set("last_login.login_method", loginMethod.String()); err != nil {
			return fmt.Errorf("failed to set login_method to the payload: %w", err)
		}

		if thirdPartyProvider.Exists() {
			if err := c.Payload().Set("last_login.third_party_provider", thirdPartyProvider.String()); err != nil {
				return fmt.Errorf("failed to set third_party_provider to the payload: %w", err)
			}
		}

		if mfaMethod.Exists() {
			if err := c.Payload().Set("last_login.mfa_method", mfaMethod.String()); err != nil {
				return fmt.Errorf("failed to set mfa_method to the payload: %w", err)
			}
		}
	}

	return nil
}
