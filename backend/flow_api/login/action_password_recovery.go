package login

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	passkeyOnboarding "github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"golang.org/x/crypto/bcrypt"
)

type PasswordRecovery struct {
	shared.Action
}

func (a PasswordRecovery) GetName() flowpilot.ActionName {
	return ActionPasswordRecovery
}

func (a PasswordRecovery) GetDescription() string {
	return "Submit a new password."
}

func (a PasswordRecovery) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	c.AddInputs(flowpilot.PasswordInput("new_password").Required(true).MinLength(deps.Cfg.Password.MinPasswordLength))
}

func (a PasswordRecovery) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	newPassword := c.Input().Get("new_password").String()

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("user_id does not exist")))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	passwordPersister := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx)

	authUserID := c.Stash().Get("user_id").String()
	passwordCredentialModel, err := passwordPersister.GetByUserID(uuid.FromStringOrNil(authUserID))
	if err != nil {
		return fmt.Errorf("failed to get password credential by user id: %w", err)
	}

	passwordCredentialModel.Password = string(hashedPassword)

	err = passwordPersister.Update(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("failed to update the password credential: %w", err)
	}

	// Decide which is the next state according to the config and user input
	if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passkeyOnboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
	}

	return c.ContinueFlow(shared.StateSuccess)
}
