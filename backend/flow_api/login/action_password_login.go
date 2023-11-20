package login

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"golang.org/x/crypto/bcrypt"
)

type PasswordLogin struct {
	shared.Action
}

func (a PasswordLogin) GetName() flowpilot.ActionName {
	return ActionPasswordLogin
}

func (a PasswordLogin) GetDescription() string {
	return "Login with a password."
}

func (a PasswordLogin) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("password").Required(true))
}

func (a PasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	var userID uuid.UUID

	if c.Stash().Get("email").Exists() {
		emailModel, err := deps.Persister.GetEmailPersister().FindByAddress(c.Stash().Get("email").String())
		if err != nil {
			return fmt.Errorf("failed to find user by email: %w", err)
		}

		if emailModel == nil {
			return a.wrongCredentialsError(c)
		}

		userID = *emailModel.UserID
	} else if c.Stash().Get("username").Exists() {
		username := c.Stash().Get("username").String()
		userModel, err := deps.Persister.GetUserPersister().GetByUsername(username)
		if err != nil {
			return fmt.Errorf("failed to find user via username: %w", err)
		}

		if userModel == nil {
			return a.wrongCredentialsError(c)
		}

		userID = userModel.ID
	} else {
		return a.wrongCredentialsError(c)
	}

	// TODO
	//if h.rateLimiter != nil {
	//	err := rate_limiter.Limit(h.rateLimiter, userId, c)
	//	if err != nil {
	//		return err
	//	}
	//}

	user, err := deps.Persister.GetUserPersister().Get(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		//err = h.auditLogger.Create(c, models.AuditLogPasswordLoginFailed, nil, fmt.Errorf("unknown user: %s", userID))
		//if err != nil {
		//	return fmt.Errorf("failed to create audit log: %w", err)
		//}
		return a.wrongCredentialsError(c)
	}

	pwBytes := []byte(c.Input().Get("password").String())
	if len(pwBytes) > 72 {
		return a.wrongCredentialsError(c)
	}

	pw, err := deps.Persister.GetPasswordCredentialPersister().GetByUserID(userID)
	if pw == nil {
		//err = h.auditLogger.Create(c, models.AuditLogPasswordLoginFailed, user, fmt.Errorf("user has no password credential"))
		//if err != nil {
		//	return fmt.Errorf("failed to create audit log: %w", err)
		//}
		return a.wrongCredentialsError(c)
	}

	if err != nil {
		return fmt.Errorf("error retrieving password credential: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(pw.Password), pwBytes); err != nil {
		//err = h.auditLogger.Create(c, models.AuditLogPasswordLoginFailed, user, fmt.Errorf("password hash not equal"))
		//if err != nil {
		//	return fmt.Errorf("failed to create audit log: %w", err)
		//}
		return a.wrongCredentialsError(c)
	}

	err = c.Stash().Set("user", user)
	if err != nil {
		return fmt.Errorf("failed to set user to stash: %w", err)
	}

	if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passkey_onboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
	}

	return c.ContinueFlow(shared.StateSuccess)
}

func (a PasswordLogin) wrongCredentialsError(c flowpilot.ExecutionContext) error {
	c.Input().SetError("password", flowpilot.ErrorValueInvalid)
	return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid.Wrap(errors.New("wrong credentials")))
}
