package passkey_onboarding

import (
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type GetWACreationOptions struct {
	cfg       config.Config
	persister persistence.Persister
	wa        *webauthnLib.WebAuthn
}

func (m GetWACreationOptions) GetName() flowpilot.ActionName {
	return shared.ActionGetWACreationOptions
}

func (m GetWACreationOptions) GetDescription() string {
	return "Get creation options to create a webauthn credential."
}

func (m GetWACreationOptions) Initialize(c flowpilot.InitializationContext) {
	webAuthnAvailable := c.Stash().Get("webauthn_available").Bool()
	if !webAuthnAvailable {
		c.SuspendAction()
	}
}

func (m GetWACreationOptions) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userId, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if !c.Stash().Get("user_id").Exists() {
		err = c.Stash().Set("user_id", userId)
		if err != nil {
			return err
		}
	} else {
		userId, err = uuid.FromString(c.Stash().Get("user_id").String())
		if err != nil {
			return err
		}
	}
	user := WebAuthnUser{
		ID:       userId,
		Email:    c.Stash().Get("email").String(),
		Username: c.Stash().Get("username").String(),
	}
	t := true
	options, sessionData, err := m.wa.BeginRegistration(
		user,
		webauthnLib.WithConveyancePreference(protocol.PreferNoAttestation),
		webauthnLib.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: &t,
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.VerificationRequired,
		}),
	)

	sessionDataModel := intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationRegistration)
	err = m.persister.GetWebauthnSessionDataPersister().Create(*sessionDataModel)
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
