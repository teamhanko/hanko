package services

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type GenerateRequestOptionsPasskeyParams struct {
	Tx   *pop.Connection
	User *models.User
}

type GenerateRequestOptionsSecurityKeyParams struct {
	Tx     *pop.Connection
	UserID uuid.UUID
}

type VerifyAssertionResponseParams struct {
	Tx                *pop.Connection
	SessionDataID     uuid.UUID
	AssertionResponse string
	IsMFA             bool
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
	GenerateRequestOptionsPasskey(GenerateRequestOptionsPasskeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error)
	GenerateRequestOptionsSecurityKey(GenerateRequestOptionsSecurityKeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error)
	VerifyAssertionResponse(VerifyAssertionResponseParams) (*models.User, error)
	GenerateCreationOptionsPasskey(GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error)
	GenerateCreationOptionsSecurityKey(GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error)
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
	ErrInvalidWebauthnCredential        = errors.New("this passkey cannot be used anymore")
	ErrInvalidWebauthnCredentialMFAOnly = errors.New("this credential can be used as a second factor security key only")
)

type webauthnService struct {
	cfg       config.Config
	persister persistence.Persister
}

func NewWebauthnService(cfg config.Config, persister persistence.Persister) WebauthnService {
	return &webauthnService{cfg: cfg, persister: persister}
}

func (s *webauthnService) generateRequestOptions(tx *pop.Connection, user webauthn.User, opts ...webauthn.LoginOption) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error) {
	var options *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	var err error
	if !reflect.ValueOf(user).IsNil() {
		options, sessionData, err = s.cfg.Webauthn.Handler.BeginLogin(user, opts...)
	} else {
		options, sessionData, err = s.cfg.Webauthn.Handler.BeginDiscoverableLogin(opts...)
	}
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

func (s *webauthnService) GenerateRequestOptionsPasskey(p GenerateRequestOptionsPasskeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error) {
	userVerificationRequirement := protocol.UserVerificationRequirement(s.cfg.Passkey.UserVerification)

	return s.generateRequestOptions(p.Tx,
		p.User,
		webauthn.WithUserVerification(userVerificationRequirement),
	)
}

func (s *webauthnService) GenerateRequestOptionsSecurityKey(p GenerateRequestOptionsSecurityKeyParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error) {
	userVerificationRequirement := protocol.UserVerificationRequirement(s.cfg.MFA.SecurityKeys.UserVerification)

	userModel, err := s.persister.GetUserPersisterWithConnection(p.Tx).Get(p.UserID)
	if err != nil || userModel == nil {
		return nil, nil, fmt.Errorf("failed to get user from db: %w", err)
	}

	credentialModels, err := s.persister.GetWebauthnCredentialPersisterWithConnection(p.Tx).GetFromUser(p.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get webauthn credentials from db: %w", err)
	}

	descriptors, err := credentialModels.GetWebauthnDescriptors()
	if err != nil {
		return nil, nil, err
	}

	return s.generateRequestOptions(p.Tx,
		userModel,
		webauthn.WithUserVerification(userVerificationRequirement),
		webauthn.WithAllowedCredentials(descriptors),
	)
}

func (s *webauthnService) VerifyAssertionResponse(p VerifyAssertionResponseParams) (*models.User, error) {
	credentialAssertionData, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(p.AssertionResponse))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	sessionDataModel, err := s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Get(p.SessionDataID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data from db: %w", err)
	}

	credentialModel, err := s.persister.GetWebauthnCredentialPersisterWithConnection(p.Tx).Get(credentialAssertionData.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webauthncredential from db: %w", err)
	}

	if credentialModel == nil {
		return nil, ErrInvalidWebauthnCredential
	}

	if !p.IsMFA && credentialModel.MFAOnly {
		return nil, ErrInvalidWebauthnCredentialMFAOnly
	}

	webAuthnUser, userModel, err := s.GetWebAuthnUser(p.Tx, *credentialModel)
	if err != nil {
		return nil, err
	}

	discoverableUserHandler := func(rawID, userHandle []byte) (webauthn.User, error) {
		return webAuthnUser, nil
	}

	sessionData := sessionDataModel.ToSessionData()
	if p.IsMFA || len(sessionData.AllowedCredentialIDs) > 0 {
		_, err = s.cfg.Webauthn.Handler.ValidateLogin(webAuthnUser, *sessionData, credentialAssertionData)
	} else {
		_, err = s.cfg.Webauthn.Handler.ValidateDiscoverableLogin(
			discoverableUserHandler,
			*sessionData,
			credentialAssertionData,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	now := time.Now().UTC()
	flags := credentialAssertionData.Response.AuthenticatorData.Flags

	credentialModel.LastUsedAt = &now
	credentialModel.BackupState = flags.HasBackupState()
	credentialModel.BackupEligible = flags.HasBackupEligible()

	err = s.persister.GetWebauthnCredentialPersisterWithConnection(p.Tx).Update(*credentialModel)
	if err != nil {
		return nil, fmt.Errorf("failed to update webauthn credential: %w", err)
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Delete(*sessionDataModel)
	if err != nil {
		return nil, fmt.Errorf("failed to delete assertion session data: %w", err)
	}

	return userModel, nil
}

func (s *webauthnService) generateCreationOptions(p GenerateCreationOptionsParams, opts ...webauthn.RegistrationOption) (*models.WebauthnSessionData, *protocol.CredentialCreation, error) {
	user := webauthnUser{id: p.UserID, email: p.Email, username: p.Username}

	userModel, err := s.persister.GetUserPersisterWithConnection(p.Tx).Get(user.id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user from db: %w", err)
	}

	// Assemble exclude list only if user already exists (i.e. the current flow is not a registration flow).
	if userModel != nil {
		credentialDescriptors, err := userModel.WebauthnCredentials.GetWebauthnDescriptors()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get credential descriptors from webauthn credentials: %w", err)
		}

		opts = append(opts, webauthn.WithExclusions(credentialDescriptors))
	}

	options, sessionData, err := s.cfg.Webauthn.Handler.BeginRegistration(user, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", err, ErrInvalidWebauthnCredential)
	}

	sessionDataModel, err := models.NewWebauthnSessionDataFrom(sessionData, models.WebauthnOperationRegistration)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new session data model instance: %w", err)
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Create(*sessionDataModel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to store session data to the db: %w", err)
	}

	return sessionDataModel, options, nil
}

func (s *webauthnService) GenerateCreationOptionsSecurityKey(p GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error) {
	requireResidentKey := false
	authenticatorSelection := protocol.AuthenticatorSelection{
		RequireResidentKey: &requireResidentKey,
		ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
		UserVerification:   protocol.UserVerificationRequirement(s.cfg.MFA.SecurityKeys.UserVerification),
	}

	authenticatorAttachment := s.cfg.MFA.SecurityKeys.AuthenticatorAttachment
	if authenticatorAttachment == "platform" || authenticatorAttachment == "cross-platform" {
		authenticatorSelection.AuthenticatorAttachment = protocol.AuthenticatorAttachment(authenticatorAttachment)
	}

	attestationPreference := protocol.ConveyancePreference(s.cfg.Passkey.AttestationPreference)

	return s.generateCreationOptions(p,
		webauthn.WithAuthenticatorSelection(authenticatorSelection),
		webauthn.WithConveyancePreference(attestationPreference),
	)
}

func (s *webauthnService) GenerateCreationOptionsPasskey(p GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error) {
	requireResidentKey := true
	authenticatorSelection := protocol.AuthenticatorSelection{
		RequireResidentKey: &requireResidentKey,
		ResidentKey:        protocol.ResidentKeyRequirementRequired,
		UserVerification:   protocol.UserVerificationRequirement(s.cfg.Passkey.UserVerification),
	}

	attestationPreference := protocol.ConveyancePreference(s.cfg.Passkey.AttestationPreference)

	return s.generateCreationOptions(p,
		webauthn.WithAuthenticatorSelection(authenticatorSelection),
		webauthn.WithConveyancePreference(attestationPreference),
	)
}

func (s *webauthnService) VerifyAttestationResponse(p VerifyAttestationResponseParams) (*webauthn.Credential, error) {
	credentialCreationData, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(p.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential creation response; %w", err)
	}

	sessionDataModel, err := s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Get(p.SessionDataID)
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

func (s *webauthnService) GetWebAuthnUser(tx *pop.Connection, credential models.WebauthnCredential) (webauthn.User, *models.User, error) {
	user, err := s.persister.GetUserPersisterWithConnection(tx).Get(credential.UserId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch user from db: %w", err)
	}
	if user == nil {
		return nil, nil, ErrInvalidWebauthnCredential
	}

	if credential.UserHandle != nil {
		return &webauthnUserWithCustomUserHandle{
			CustomUserHandle: []byte(credential.UserHandle.Handle),
			User:             *user,
		}, user, nil
	}

	return user, user, err
}

type webauthnUserWithCustomUserHandle struct {
	models.User
	CustomUserHandle []byte
}

func (u *webauthnUserWithCustomUserHandle) WebAuthnID() []byte {
	return u.CustomUserHandle
}
