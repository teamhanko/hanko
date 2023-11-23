package services

import (
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

type GenerateRequestOptionsParams struct {
	Tx *pop.Connection
}

type VerifyAssertionResponseParams struct {
	Tx                *pop.Connection
	SessionDataID     uuid.UUID
	AssertionResponse string
}

type GenerateCreationOptionsParams struct {
	Tx       *pop.Connection
	UserID   uuid.UUID
	Email    string
	Username string
}

type VerifyAttestationResponseParams struct {
	Tx            *pop.Connection
	SessionDataID uuid.UUID
	PublicKey     string
	UserID        uuid.UUID
	Email         string
	Username      string
}

type WebauthnService interface {
	GenerateRequestOptions(GenerateRequestOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error)
	VerifyAssertionResponse(VerifyAssertionResponseParams) error
	GenerateCreationOptions(GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error)
	VerifyAttestationResponse(VerifyAttestationResponseParams) (*webauthn.Credential, error)
}

type webauthnUser struct {
	id       uuid.UUID
	email    string
	username string
}

func (user webauthnUser) WebAuthnID() []byte {
	return user.id.Bytes()
}

func (user webauthnUser) WebAuthnName() string {
	if len(user.email) > 0 {
		return user.email
	}

	return user.username
}

func (user webauthnUser) WebAuthnDisplayName() string {
	if len(user.username) > 0 {
		return user.username
	}

	return user.email
}

func (user webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return nil
}

func (user webauthnUser) WebAuthnIcon() string {
	return ""
}

type Error struct {
	Details string
}

func (e Error) Error() string {
	return e.Details
}

func (e Error) Wrap(err error) error {
	return &Error{Details: fmt.Sprintf("%s: %s", e.Details, err.Error())}
}

var (
	ErrInvalidWebauthnCredential = &Error{Details: "this passcode cannot be used anymore"}
)

type webauthnService struct {
	cfg       config.Config
	persister persistence.Persister
}

func NewWebauthnService(cfg config.Config, persister persistence.Persister) WebauthnService {
	return &webauthnService{cfg: cfg, persister: persister}
}

func (s *webauthnService) GenerateRequestOptions(p GenerateRequestOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialAssertion, error) {
	userVerificationRequirement := protocol.UserVerificationRequirement(s.cfg.Webauthn.UserVerification)
	options, sessionData, err := s.cfg.Webauthn.Handler.BeginDiscoverableLogin(
		webauthn.WithUserVerification(userVerificationRequirement),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create webauthn assertion options for discoverable login: %w", err)
	}

	webAuthnSessionDataModel, err := models.NewWebauthnSessionDataFrom(sessionData, models.WebauthnOperationAuthentication)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate a new webauthn session data model: %w", err)
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Create(*webAuthnSessionDataModel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	}

	return webAuthnSessionDataModel, options, nil
}

func (s *webauthnService) VerifyAssertionResponse(p VerifyAssertionResponseParams) error {
	credentialAssertionData, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(p.AssertionResponse))
	if err != nil {
		return ErrInvalidWebauthnCredential.Wrap(err)
	}

	sessionDataModel, err := s.persister.GetWebauthnSessionDataPersister().Get(p.SessionDataID)
	if err != nil {
		return fmt.Errorf("failed to get session data from db: %w", err)
	}

	userID, err := uuid.FromBytes(credentialAssertionData.Response.UserHandle)
	if err != nil {
		return fmt.Errorf("failed to parse user id from user handle: %w", err)
	}

	userModel, err := s.persister.GetUserPersister().Get(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user from db: %w", err)
	}

	if userModel == nil {
		return ErrInvalidWebauthnCredential.Wrap(errors.New("user does not exist"))
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
		return ErrInvalidWebauthnCredential.Wrap(err)
	}

	if credentialModel := userModel.GetWebauthnCredentialById(credential.ID); credentialModel != nil {
		now := time.Now().UTC()
		flags := credentialAssertionData.Response.AuthenticatorData.Flags

		credentialModel.LastUsedAt = &now
		credentialModel.BackupState = flags.HasBackupState()
		credentialModel.BackupEligible = flags.HasBackupEligible()

		err = s.persister.GetWebauthnCredentialPersisterWithConnection(p.Tx).Update(*credentialModel)
		if err != nil {
			return fmt.Errorf("failed to update webauthn credential: %w", err)
		}
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Delete(*sessionDataModel)
	if err != nil {
		return fmt.Errorf("failed to delete assertion session data: %w", err)
	}

	return nil
}

func (s *webauthnService) GenerateCreationOptions(p GenerateCreationOptionsParams) (*models.WebauthnSessionData, *protocol.CredentialCreation, error) {
	user := webauthnUser{id: p.UserID, email: p.Email, username: p.Username}

	requireResidentKey := true
	authenticatorSelection := protocol.AuthenticatorSelection{
		RequireResidentKey: &requireResidentKey,
		ResidentKey:        protocol.ResidentKeyRequirementRequired,
		UserVerification:   protocol.VerificationRequired,
	}

	options, sessionData, err := s.cfg.Webauthn.Handler.BeginRegistration(
		user,
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
		webauthn.WithAuthenticatorSelection(authenticatorSelection),
	)
	if err != nil {
		return nil, nil, ErrInvalidWebauthnCredential.Wrap(err)
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
		return nil, ErrInvalidWebauthnCredential.Wrap(err)
	}

	err = s.persister.GetWebauthnSessionDataPersisterWithConnection(p.Tx).Delete(*sessionDataModel)
	if err != nil {
		return nil, fmt.Errorf("failed to delete webauthn session data: %w", err)
	}

	return credential, nil
}
