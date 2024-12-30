package shared

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"time"
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

	var generateJWTOptions []session.JWTOptions
	if userModel.Username != nil {
		generateJWTOptions = append(generateJWTOptions, session.WithValue("username", userModel.Username.Username))
	}

	signedSessionToken, rawToken, err := deps.SessionManager.GenerateJWT(userId, emailDTO, generateJWTOptions...)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	claims, err := dto.GetClaimsFromToken(rawToken)
	if err != nil {
		return fmt.Errorf("failed to get token claims: %w", err)
	}

	err = c.Payload().Set("claims", claims)
	if err != nil {
		return fmt.Errorf("failed to set token claims to payload: %w", err)
	}

	activeSessions, err := deps.Persister.GetSessionPersisterWithConnection(deps.Tx).ListActive(userId)
	if err != nil {
		return fmt.Errorf("failed to list active sessions: %w", err)
	}

	// remove all server side sessions that exceed the limit
	if len(activeSessions) >= deps.Cfg.Session.Limit {
		for i := deps.Cfg.Session.Limit - 1; i < len(activeSessions); i++ {
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
		CreatedAt: rawToken.IssuedAt(),
		UpdatedAt: rawToken.IssuedAt(),
		ExpiresAt: &expirationTime,
		LastUsed:  rawToken.IssuedAt(),
	}

	if deps.Cfg.Session.AcquireIPAddress {
		sessionModel.IpAddress = nulls.NewString(deps.HttpContext.RealIP())
	}

	if deps.Cfg.Session.AcquireUserAgent {
		sessionModel.UserAgent = nulls.NewString(deps.HttpContext.Request().UserAgent())
	}

	err = deps.Persister.GetSessionPersisterWithConnection(deps.Tx).Create(sessionModel)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	rememberMeSelected := c.Stash().Get(StashPathRememberMeSelected).Bool()
	cookie, err := deps.SessionManager.GenerateCookie(signedSessionToken)
	if err != nil {
		return fmt.Errorf("failed to generate auth cookie, %w", err)
	}

	lifespan, err := time.ParseDuration(deps.Cfg.Session.Lifespan)
	if err != nil {
		return fmt.Errorf("failed to parse session lifespan: %w", err)
	}

	sessionRetention := "persistent"
	if deps.Cfg.Session.Cookie.Retention == "session" ||
		(deps.Cfg.Session.Cookie.Retention == "prompt" && !rememberMeSelected) {
		// Issue a session cookie.
		cookie.MaxAge = 0
		sessionRetention = "session"
	}

	deps.HttpContext.Response().Header().Set("X-Session-Lifetime", fmt.Sprintf("%d", int(lifespan.Seconds())))
	deps.HttpContext.Response().Header().Set("X-Session-Retention", fmt.Sprintf("%s", sessionRetention))

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
