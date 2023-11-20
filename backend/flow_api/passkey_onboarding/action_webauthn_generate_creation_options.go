package passkey_onboarding

import (
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnGenerateCreationOptions struct {
	shared.Action
}

func (a WebauthnGenerateCreationOptions) GetName() flowpilot.ActionName {
	return ActionWebauthnGenerateCreationOptions
}

func (a WebauthnGenerateCreationOptions) GetDescription() string {
	return "Get creation options to create a webauthn credential."
}

func (a WebauthnGenerateCreationOptions) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("webauthn_available").Bool() {
		c.SuspendAction()
	}
}

func (a WebauthnGenerateCreationOptions) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userId, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate a new user id: %w", err)
	}

	if !c.Stash().Get("user_id").Exists() {
		err = c.Stash().Set("user_id", userId)
		if err != nil {
			return fmt.Errorf("failed to sett user id to the stash: %w", err)
		}
	} else {
		userId, err = uuid.FromString(c.Stash().Get("user_id").String())
		if err != nil {
			return fmt.Errorf("failed to parse stashed user id as a uuid: %w", err)
		}
	}

	user := WebAuthnUser{
		ID:       userId,
		Email:    c.Stash().Get("email").String(),
		Username: c.Stash().Get("username").String(),
	}

	requireResidentKey := true

	options, sessionData, err := deps.Cfg.Webauthn.Handler.BeginRegistration(
		user,
		webauthnLib.WithConveyancePreference(protocol.PreferNoAttestation),
		webauthnLib.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: &requireResidentKey,
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.VerificationRequired,
		}),
	)

	sessionDataModel := intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationRegistration)
	err = deps.Persister.GetWebauthnSessionDataPersister().Create(*sessionDataModel)
	if err != nil {
		return err
	}

	err = c.Stash().Set("webauthn_session_data_id", sessionDataModel.ID)
	if err != nil {
		return err
	}

	err = c.Payload().Set("creationOptions", options)
	if err != nil {
		return err
	}

	return c.ContinueFlow(StateOnboardingVerifyPasskeyAttestation)
}

type WebAuthnUser struct {
	ID       uuid.UUID
	Email    string
	Username string
}

func (u WebAuthnUser) WebAuthnID() []byte {
	return u.ID.Bytes()
}

func (u WebAuthnUser) WebAuthnName() string {
	if u.Email != "" {
		return u.Email
	}

	return u.Username
}

func (u WebAuthnUser) WebAuthnDisplayName() string {
	if u.Username != "" {
		return u.Username
	}

	return u.Email
}

func (u WebAuthnUser) WebAuthnCredentials() []webauthnLib.Credential {
	// TODO: when we use this action also in the profile or in the login flow, then we should/must add the users credentials here.
	return nil
}

func (u WebAuthnUser) WebAuthnIcon() string {
	return ""
}
