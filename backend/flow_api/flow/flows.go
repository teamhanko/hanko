package flow

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/credential_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login_method_chooser"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login_password"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/profile"
	"github.com/teamhanko/hanko/backend/flow_api/flow/registration"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

var CapabilitiesSubFlow = flowpilot.NewSubFlow("capabilities").
	State(shared.StatePreflight, capabilities.RegisterClientCapabilities{}).
	MustBuild()

var LoginMethodChooserSubFlow = flowpilot.NewSubFlow("login_method_chooser").
	State(shared.StateLoginMethodChooser,
		login_method_chooser.ContinueToPasswordLogin{},
		login_method_chooser.ContinueToPasscodeConfirmation{},
		shared.Back{},
	).
	SubFlows(LoginPasswordSubFlow, PasscodeSubFlow).
	MustBuild()

var LoginPasswordSubFlow = flowpilot.NewSubFlow("login_password").
	State(shared.StateLoginPassword,
		login_password.PasswordLogin{},
		login_password.ContinueToPasscodeConfirmationRecovery{},
		shared.Back{},
	).
	State(shared.StateLoginPasswordRecovery, login_password.PasswordRecovery{}).
	SubFlows(PasscodeSubFlow).
	MustBuild()

var PasscodeSubFlow = flowpilot.NewSubFlow("passcode").
	State(shared.StatePasscodeConfirmation, passcode.VerifyPasscode{}, passcode.ReSendPasscode{}, shared.Back{}).
	BeforeState(shared.StatePasscodeConfirmation, passcode.SendPasscode{}).
	MustBuild()

var CredentialOnboardingSubFlow = flowpilot.NewSubFlow("credential_onboarding").
	State(shared.StateCredentialOnboardingChooser,
		credential_onboarding.ContinueToPasskey{},
		credential_onboarding.ContinueToPassword{},
		credential_onboarding.SkipCredentialOnboardingMethodChooser{},
		credential_onboarding.BackCredentialOnboardingMethodChooser{}).
	State(shared.StateOnboardingCreatePasskey,
		credential_onboarding.WebauthnGenerateCreationOptions{},
		credential_onboarding.SkipPasskey{},
		credential_onboarding.Back{}).
	State(shared.StateOnboardingVerifyPasskeyAttestation,
		credential_onboarding.WebauthnVerifyAttestationResponse{},
		shared.Back{}).
	State(shared.StatePasswordCreation,
		credential_onboarding.RegisterPassword{},
		credential_onboarding.Back{},
		credential_onboarding.SkipPassword{}).
	MustBuild()

var LoginFlow = flowpilot.NewFlow("/login").
	State(shared.StateLoginInit,
		login.ContinueWithLoginIdentifier{},
		login.WebauthnGenerateRequestOptions{},
		login.WebauthnVerifyAssertionResponse{},
		shared.ThirdPartyOAuth{}).
	State(shared.StateThirdPartyOAuth, shared.ExchangeToken{}).
	State(shared.StateLoginPasskey, login.WebauthnVerifyAssertionResponse{}, shared.Back{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	InitialState(shared.StatePreflight, shared.StateLoginInit).
	ErrorState(shared.StateError).
	BeforeState(shared.StateLoginInit, login.WebauthnGenerateRequestOptionsForConditionalUi{}).
	BeforeState(shared.StateSuccess, shared.IssueSession{}).
	BeforeState(shared.StatePasscodeConfirmation, login.SelectPasscodeTemplate{}).
	AfterState(shared.StateOnboardingVerifyPasskeyAttestation, shared.WebauthnCredentialSave{}).
	AfterState(shared.StatePasscodeConfirmation, shared.EmailPersistVerifiedStatus{}).
	SubFlows(CapabilitiesSubFlow, PasscodeSubFlow, LoginMethodChooserSubFlow, LoginPasswordSubFlow, CredentialOnboardingSubFlow).
	TTL(10 * time.Minute).
	Debug(true)

var RegistrationFlow = flowpilot.NewFlow("/registration").
	State(shared.StateRegistrationInit, registration.RegisterLoginIdentifier{}, shared.ThirdPartyOAuth{}).
	State(shared.StateThirdPartyOAuth, shared.ExchangeToken{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	InitialState(shared.StatePreflight, shared.StateRegistrationInit).
	ErrorState(shared.StateError).
	BeforeState(shared.StateSuccess, registration.CreateUser{}, shared.IssueSession{}).
	SubFlows(CapabilitiesSubFlow, PasscodeSubFlow, CredentialOnboardingSubFlow).
	TTL(10 * time.Minute).
	Debug(true)

var ProfileFlow = flowpilot.NewFlow("/profile").
	State(shared.StateProfileInit,
		profile.AccountDelete{},
		profile.EmailCreate{},
		profile.EmailDelete{},
		profile.EmailSetPrimary{},
		profile.EmailVerify{},
		profile.PasswordSet{},
		profile.PasswordDelete{},
		profile.UsernameSet{},
		profile.WebauthnCredentialRename{},
		profile.WebauthnCredentialCreate{},
		profile.WebauthnCredentialDelete{},
	).
	State(shared.StateProfileWebauthnCredentialVerification, profile.WebauthnVerifyAttestationResponse{}, shared.Back{}).
	State(shared.StateProfileAccountDeleted).
	InitialState(shared.StatePreflight, shared.StateProfileInit).
	ErrorState(shared.StateError).
	BeforeEachAction(profile.RefreshSessionUser{}).
	BeforeState(shared.StateProfileInit, profile.GetProfileData{}).
	AfterEachAction(profile.RefreshSessionUser{}).
	AfterState(shared.StateProfileWebauthnCredentialVerification, shared.WebauthnCredentialSave{}).
	AfterState(shared.StatePasscodeConfirmation, shared.EmailPersistVerifiedStatus{}).
	SubFlows(CapabilitiesSubFlow, PasscodeSubFlow).
	TTL(10 * time.Minute).
	Debug(true)
