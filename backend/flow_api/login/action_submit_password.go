package login

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"golang.org/x/crypto/bcrypt"
)

type SubmitPassword struct {
	cfg       config.Config
	persister persistence.Persister
}

func (a SubmitPassword) GetName() flowpilot.ActionName {
	return shared.ActionSubmitPassword
}

func (a SubmitPassword) GetDescription() string {
	return "Login with a password."
}

func (a SubmitPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("password").Required(true))
}

func (a SubmitPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	var userID uuid.UUID

	if c.Stash().Get("email").Exists() {
		emailModel, err := a.persister.GetEmailPersister().FindByAddress(c.Stash().Get("email").String())
		if err != nil {
			return fmt.Errorf("failed to find user by email: %w", err)
		}

		if emailModel == nil {
			return a.wrongCredentialsError(c)
		}

		userID = *emailModel.UserID
	} else if c.Stash().Get("username").Exists() {
		username := c.Stash().Get("username").String()
		userModel, err := a.persister.GetUserPersister().GetByUsername(username)
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

	user, err := a.persister.GetUserPersister().Get(userID)
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

	pw, err := a.persister.GetPasswordCredentialPersister().GetByUserID(userID)
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

	if a.cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passkey_onboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
	}

	return c.ContinueFlow(shared.StateSuccess)
}

func (a SubmitPassword) wrongCredentialsError(c flowpilot.ExecutionContext) error {
	c.Input().SetError("password", flowpilot.ErrorValueInvalid)
	return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid.Wrap(errors.New("wrong credentials")))
}
