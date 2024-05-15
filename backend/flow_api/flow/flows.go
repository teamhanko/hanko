package flow

import (
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login_method_chooser"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login_password"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/profile"
	"github.com/teamhanko/hanko/backend/flow_api/flow/register_password"
	"github.com/teamhanko/hanko/backend/flow_api/flow/registration"
	"github.com/teamhanko/hanko/backend/flow_api/flow/registration_method_chooser"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

var CapabilitiesSubFlow = flowpilot.NewSubFlow("capabilities").
	State(constants.StatePreflight, capabilities.RegisterClientCapabilities{}).
	MustBuild()

var LoginMethodChooserSubFlow = flowpilot.NewSubFlow("login_method_chooser").
	State(constants.StateLoginMethodChooser,
		login_method_chooser.ContinueToPasswordLogin{},
		login_method_chooser.ContinueToPasscodeConfirmation{},
		shared.Back{},
	).
	SubFlows(LoginPasswordSubFlow, PasscodeSubFlow).
	MustBuild()

var LoginPasswordSubFlow = flowpilot.NewSubFlow("login_password").
	State(constants.StateLoginPassword,
		login_password.PasswordLogin{},
		login_password.ContinueToPasscodeConfirmationRecovery{},
		shared.Back{},
	).
	State(constants.StateLoginPasswordRecovery, login_password.PasswordRecovery{}).
	SubFlows(PasscodeSubFlow).
	MustBuild()

var PasscodeSubFlow = flowpilot.NewSubFlow("passcode").
	State(constants.StatePasscodeConfirmation, passcode.VerifyPasscode{}, passcode.ReSendPasscode{}, shared.Back{}).
	BeforeState(constants.StatePasscodeConfirmation, passcode.SendPasscode{}).
	MustBuild()

var PasskeyOnboardingSubFlow = flowpilot.NewSubFlow("passkey_onboarding").
	State(constants.StateOnboardingCreatePasskey, passkey_onboarding.WebauthnGenerateCreationOptions{}, passkey_onboarding.Skip{}, passkey_onboarding.Back{}).
	State(constants.StateOnboardingVerifyPasskeyAttestation, passkey_onboarding.WebauthnVerifyAttestationResponse{}, shared.Back{}).
	// SubFlows(RegisterPasswordSubFlow).
	MustBuild()

var RegisterPasswordSubFlow = flowpilot.NewSubFlow("register_password").
	State(constants.StatePasswordCreation, register_password.RegisterPassword{}, register_password.Back{}, register_password.Skip{}).
	// SubFlows(PasskeyOnboardingSubFlow).
	MustBuild()

var ConditionalOnboarding = flowpilot.NewSubFlow("conditional_onboarding").
	State(constants.StateOnboardingCreatePasskeyConditional, passkey_onboarding.WebauthnGenerateCreationOptions{}, passkey_onboarding.Skip{}, passkey_onboarding.Back{}).
	State(constants.StateOnboardingVerifyPasskeyAttestation, passkey_onboarding.WebauthnVerifyAttestationResponse{}, shared.Back{}).
	State(constants.StatePasswordCreationConditional, register_password.RegisterPassword{}, register_password.Back{}, register_password.Skip{}).
	MustBuild()

var RegistrationMethodChooserSubFlow = flowpilot.NewSubFlow("registration_method_chooser").
	State(constants.StateRegistrationMethodChooser,
		registration_method_chooser.ContinueToPasskeyCreation{},
		registration_method_chooser.ContinueToPasswordRegistration{},
		registration_method_chooser.Back{},
		registration_method_chooser.Skip{}).
	SubFlows(PasskeyOnboardingSubFlow, RegisterPasswordSubFlow).
	MustBuild()

var LoginFlow = flowpilot.NewFlow("/login").
	State(constants.StateLoginInit,
		login.ContinueWithLoginIdentifier{},
		login.WebauthnGenerateRequestOptions{},
		login.WebauthnVerifyAssertionResponse{},
		shared.ThirdPartyOAuth{}).
	BeforeState(constants.StateLoginInit, login.WebauthnGenerateRequestOptionsForConditionalUi{}).
	State(constants.StateThirdPartyOAuth, shared.ExchangeToken{}).
	State(constants.StateLoginPasskey, login.WebauthnVerifyAssertionResponse{}, shared.Back{}).
	BeforeState(constants.StateSuccess, shared.IssueSession{}).
	State(constants.StateSuccess).
	State(constants.StateError).
	SubFlows(CapabilitiesSubFlow, PasskeyOnboardingSubFlow, PasscodeSubFlow, LoginMethodChooserSubFlow, LoginPasswordSubFlow, RegisterPasswordSubFlow, ConditionalOnboarding).
	AfterState(constants.StateOnboardingVerifyPasskeyAttestation, shared.WebauthnCredentialSave{}).
	InitialState(constants.StatePreflight, constants.StateLoginInit).
	BeforeState(constants.StatePasscodeConfirmation, login.SelectPasscodeTemplate{}).
	AfterState(constants.StatePasscodeConfirmation, shared.EmailPersistVerifiedStatus{}).
	ErrorState(constants.StateError).
	TTL(10 * time.Minute)

var RegistrationFlow = flowpilot.NewFlow("/registration").
	State(constants.StateRegistrationInit, registration.RegisterLoginIdentifier{}, shared.ThirdPartyOAuth{}).
	State(constants.StateThirdPartyOAuth, shared.ExchangeToken{}).
	BeforeState(constants.StateSuccess, registration.CreateUser{}, shared.IssueSession{}).
	State(constants.StateSuccess).
	State(constants.StateError).
	SubFlows(CapabilitiesSubFlow, RegistrationMethodChooserSubFlow, PasscodeSubFlow, PasskeyOnboardingSubFlow, RegisterPasswordSubFlow).
	InitialState(constants.StatePreflight, constants.StateRegistrationInit).
	ErrorState(constants.StateError).
	TTL(10 * time.Minute).
	Debug(true)

var ProfileFlow = flowpilot.NewFlow("/profile").
	InitialState(constants.StatePreflight, constants.StateProfileInit).
	BeforeEachAction(profile.RefreshSessionUser{}).
	AfterEachAction(profile.RefreshSessionUser{}).
	BeforeState(constants.StateProfileInit, profile.GetProfileData{}).
	State(constants.StateProfileInit,
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
	State(constants.StateProfileWebauthnCredentialVerification, profile.WebauthnVerifyAttestationResponse{}, shared.Back{}).
	AfterState(constants.StateProfileWebauthnCredentialVerification, shared.WebauthnCredentialSave{}).
	AfterState(constants.StatePasscodeConfirmation, shared.EmailPersistVerifiedStatus{}).
	State(constants.StateProfileAccountDeleted).
	ErrorState(constants.StateError).
	SubFlows(CapabilitiesSubFlow, PasscodeSubFlow).
	TTL(10 * time.Minute).
	Debug(true)
