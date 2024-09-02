package services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"strings"
	"time"
)

type GenerateRequestOptionsPasskeyParams struct {
	Tx *pop.Connection
}

type GenerateRequestOptionsSecurityKeyParams struct {
	Tx     *pop.Connection
	UserID uuid.UUID
}

type VerifyAssertionResponseParams struct {
	Tx                *pop.Connection
	SessionDataID     uuid.UUID
	AssertionResponse string
}

type GenerateCreationOptionsParams struct {
	Tx       *pop.Connection
	UserID   uuid.UUID
	Email    *string
	Username *string
}

type VerifyAttestationResponseParams struct {
	Tx            *pop.Connection
	SessionDataID uuid.UUID
	PublicKey     string
	UserID        uuid.UUID
	Email         *string
	Username      *string
}

type WebauthnService interface {
	GenerateRequestOptionsPasscode(GenerateRequestOptionsPasskeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error)
	GenerateRequestOptionsSecurityKey(GenerateRequestOptionsSecurityKeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error)
	VerifyAssertionResponse(VerifyAssertionResponseParams) (*models.User, error)
	GenerateCreationOptions(GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error)
	VerifyAttestationResponse(VerifyAttestationResponseParams) (*webauthn.Credential, error)
}

type webauthnUser struct {
	id       uuid.UUID
	email    *string
	username *string
}

func (user webauthnUser) WebAuthnID() []byte {
	return user.id.Bytes()
}

func (user webauthnUser) WebAuthnName() string {
	if user.email != nil && len(*user.email) > 0 {
		return *user.email
	}

	if user.username != nil {
		return *user.username
	}

	return ""
}

func (user webauthnUser) WebAuthnDisplayName() string {
	if user.username != nil && len(*user.username) > 0 {
		return *user.username
	}

	if user.email != nil {
		return *user.email
	}

	return ""
}

func (user webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return nil
}

func (user webauthnUser) WebAuthnIcon() string {
	return ""
}

var (
	ErrInvalidWebauthnCredential = errors.New("this passkey cannot be used anymore")
)

type webauthnService struct {
	cfg       config.Config
	persister persistence.Persister
}

func NewWebauthnService(cfg config.Config, persister persistence.Persister) WebauthnService {
	return &webauthnService{cfg: cfg, persister: persister}
}

func (s *webauthnService) generateRequestOptions(tx *pop.Connection, opts ...webauthn.LoginOption) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error) {
	options, sessionData, err := s.cfg.Webauthn.Handler.BeginDiscoverableLogin(opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create webauthn assertion options for discoverable login: %w", err)
	}

	webAuthnSessionDataModel, err := models.NewWebauthnSessionDataFrom(sessionData, models.WebauthnOperationAuthentication)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate a new webauthn session data model: %w", err)
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(tx).Create(*webAuthnSessionDataModel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	}

	return webAuthnSessionDataModel, options, nil
}

func (s *webauthnService) GenerateRequestOptionsPasscode(p GenerateRequestOptionsPasskeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error) {
	userVerificationRequirement := protocol.UserVerificationRequirement(s.cfg.Passkey.UserVerification)

	return s.generateRequestOptions(p.Tx,
		webauthn.WithUserVerification(userVerificationRequirement),
	)
}

func (s *webauthnService) GenerateRequestOptionsSecurityKey(p GenerateRequestOptionsSecurityKeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error) {
	userVerificationRequirement := protocol.UserVerificationRequirement(s.cfg.MFA.SecurityKeys.UserVerification)

	credentialModels, err := s.persister.GetWebauthnCredentialPersisterWithConnection(p.Tx).GetFromUser(p.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get webauthn credentials from db: %w", err)
	}

	return s.generateRequestOptions(p.Tx,
		webauthn.WithUserVerification(userVerificationRequirement),
		webauthn.WithAllowedCredentials(credentialModels.GetWebauthnDescriptors()),
	)
}

func (s *webauthnService) VerifyAssertionResponse(p VerifyAssertionResponseParams) (*models.User, error) {
	credentialAssertionData, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(p.AssertionResponse))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	sessionDataModel, err := s.persister.GetWebauthnSessionDataPersister().Get(p.SessionDataID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data from db: %w", err)
	}

	userID, err := uuid.FromBytes(credentialAssertionData.Response.UserHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id from user handle: %w", err)
	}

	userModel, err := s.persister.GetUserPersister().Get(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from db: %w", err)
	}

	if userModel == nil {
		return nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	discoverableUserHandler := func(rawID, userHandle []byte) (webauthn.User, error) {
		return userModel, nil
	}

	sessionData := sessionDataModel.ToSessionData()

	credential, err := s.cfg.Webauthn.Handler.ValidateDiscoverableLogin(
		discoverableUserHandler,
		*sessionData,
		credentialAssertionData,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	encodedCredentialId := base64.RawURLEncoding.EncodeToString(credential.ID)
	if credentialModel := userModel.GetWebauthnCredentialById(encodedCredentialId); credentialModel != nil {
		now := time.Now().UTC()
		flags := credentialAssertionData.Response.AuthenticatorData.Flags

		credentialModel.LastUsedAt = &now
		credentialModel.BackupState = flags.HasBackupState()
		credentialModel.BackupEligible = flags.HasBackupEligible()

		err = s.persister.GetWebauthnCredentialPersisterWithConnection(p.Tx).Update(*credentialModel)
		if err != nil {
			return nil, fmt.Errorf("failed to update webauthn credential: %w", err)
		}
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Delete(*sessionDataModel)
	if err != nil {
		return nil, fmt.Errorf("failed to delete assertion session data: %w", err)
	}

	return userModel, nil
}

func (s *webauthnService) GenerateCreationOptions(p GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error) {
	user := webauthnUser{id: p.UserID, email: p.Email, username: p.Username}

	requireResidentKey := true
	authenticatorSelection := protocol.AuthenticatorSelection{
		RequireResidentKey: &requireResidentKey,
		ResidentKey:        protocol.ResidentKeyRequirementRequired,
		UserVerification:   protocol.VerificationRequired,
	}

	attestationPreference := protocol.ConveyancePreference(s.cfg.Passkey.AttestationPreference)
	options, sessionData, err := s.cfg.Webauthn.Handler.BeginRegistration(
		user,
		webauthn.WithConveyancePreference(attestationPreference),
		webauthn.WithAuthenticatorSelection(authenticatorSelection),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	sessionDataModel, err := models.NewWebauthnSessionDataFrom(sessionData, models.WebauthnOperationRegistration)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new session data model instance: %w", err)
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Create(*sessionDataModel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to store session data to the db: %W", err)
	}

	return sessionDataModel, options, nil
}

func (s *webauthnService) VerifyAttestationResponse(p VerifyAttestationResponseParams) (*webauthn.Credential, error) {
	credentialCreationData, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(p.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential creation response; %w", err)
	}

	sessionDataModel, err := s.persister.GetWebauthnSessionDataPersister().Get(p.SessionDataID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data from db: %w", err)
	}

	user := webauthnUser{id: p.UserID, email: p.Email, username: p.Username}

	sessionData := sessionDataModel.ToSessionData()

	credential, err := s.cfg.Webauthn.Handler.CreateCredential(
		user,
		*sessionData,
		credentialCreationData,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Delete(*sessionDataModel)
	if err != nil {
		return nil, fmt.Errorf("failed to delete webauthn session data: %w", err)
	}

	return credential, nil
}
