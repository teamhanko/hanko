package flow

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/credential_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/credential_usage"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login"
	"github.com/teamhanko/hanko/backend/flow_api/flow/profile"
	"github.com/teamhanko/hanko/backend/flow_api/flow/registration"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/flow/user_details"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

var CapabilitiesSubFlow = flowpilot.NewSubFlow(shared.FlowCapabilities).
	State(shared.StatePreflight, capabilities.RegisterClientCapabilities{}).
	MustBuild()

var CredentialUsageSubFlow = flowpilot.NewSubFlow(shared.FlowCredentialUsage).
	State(shared.StateLoginMethodChooser,
		credential_usage.ContinueToPasswordLogin{},
		credential_usage.ContinueToPasscodeConfirmation{},
		shared.Back{},
	).
	State(shared.StateLoginPassword,
		credential_usage.PasswordLogin{},
		credential_usage.ContinueToPasscodeConfirmationRecovery{},
		shared.Back{},
	).
	State(shared.StateLoginPasswordRecovery,
		credential_usage.PasswordRecovery{}).
	State(shared.StatePasscodeConfirmation,
		credential_usage.VerifyPasscode{},
		credential_usage.ReSendPasscode{},
		shared.Back{}).
	BeforeState(shared.StatePasscodeConfirmation,
		credential_usage.SendPasscode{}).
	MustBuild()

var CredentialOnboardingSubFlow = flowpilot.NewSubFlow(shared.FlowCredentialOnboarding).
	State(shared.StateCredentialOnboardingChooser,
		credential_onboarding.ContinueToPasskey{},
		credential_onboarding.ContinueToPassword{},
		credential_onboarding.SkipCredentialOnboardingMethodChooser{},
		shared.Back{}).
	State(shared.StateOnboardingCreatePasskey,
		credential_onboarding.WebauthnGenerateCreationOptions{},
		credential_onboarding.SkipPasskey{},
		shared.Back{}).
	State(shared.StateOnboardingVerifyPasskeyAttestation,
		credential_onboarding.WebauthnVerifyAttestationResponse{},
		shared.Back{}).
	State(shared.StatePasswordCreation,
		credential_onboarding.RegisterPassword{},
		credential_onboarding.SkipPassword{},
		shared.Back{}).
	MustBuild()

var UserDetailsSubFlow = flowpilot.NewSubFlow(shared.FlowUserDetails).
	State(shared.StateOnboardingUsername,
		user_details.UsernameSet{},
		user_details.SkipUsername{}).
	State(shared.StateOnboardingEmail,
		user_details.EmailAddressSet{},
		user_details.SkipEmail{}).
	MustBuild()

var LoginFlow = flowpilot.NewFlow(shared.FlowLogin).
	State(shared.StateLoginInit,
		login.ContinueWithLoginIdentifier{},
		login.WebauthnGenerateRequestOptions{},
		login.WebauthnVerifyAssertionResponse{},
		shared.ThirdPartyOAuth{}).
	State(shared.StateThirdParty,
		shared.ExchangeToken{}).
	State(shared.StateLoginPasskey,
		login.WebauthnVerifyAssertionResponse{},
		shared.Back{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	InitialState(shared.StatePreflight, shared.StateLoginInit).
	ErrorState(shared.StateError).
	BeforeState(shared.StateLoginInit,
		login.WebauthnGenerateRequestOptionsForConditionalUi{}).
	BeforeState(shared.StateSuccess,
		shared.IssueSession{}).
	AfterState(shared.StateOnboardingVerifyPasskeyAttestation,
		shared.WebauthnCredentialSave{}).
	AfterState(shared.StatePasscodeConfirmation,
		shared.EmailPersistVerifiedStatus{}).
	AfterState(shared.StatePasswordCreation,
		shared.PasswordSave{}).
	AfterFlow(shared.FlowCredentialUsage, login.ScheduleOnboardingStates{}).
	SubFlows(
		CapabilitiesSubFlow,
		CredentialUsageSubFlow,
		CredentialOnboardingSubFlow,
		UserDetailsSubFlow).
	TTL(24 * time.Hour).
	Debug(true)

var RegistrationFlow = flowpilot.NewFlow(shared.FlowRegistration).
	State(shared.StateRegistrationInit,
		registration.RegisterLoginIdentifier{},
		shared.ThirdPartyOAuth{}).
	State(shared.StateThirdParty,
		shared.ExchangeToken{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	InitialState(shared.StatePreflight,
		shared.StateRegistrationInit).
	ErrorState(shared.StateError).
	BeforeState(shared.StateSuccess,
		registration.CreateUser{},
		shared.IssueSession{}).
	SubFlows(
		CapabilitiesSubFlow,
		CredentialUsageSubFlow,
		CredentialOnboardingSubFlow,
		UserDetailsSubFlow).
	TTL(24 * time.Hour).
	Debug(true)

var ProfileFlow = flowpilot.NewFlow(shared.FlowProfile).
	State(shared.StateProfileInit,
		profile.AccountDelete{},
		profile.EmailCreate{},
		profile.EmailDelete{},
		profile.EmailSetPrimary{},
		profile.EmailVerify{},
		profile.PasswordCreate{},
		profile.PasswordUpdate{},
		profile.PasswordDelete{},
		profile.UsernameSet{},
		profile.UsernameDelete{},
		profile.WebauthnCredentialRename{},
		profile.WebauthnCredentialCreate{},
		profile.WebauthnCredentialDelete{},
	).
	State(shared.StateProfileWebauthnCredentialVerification,
		profile.WebauthnVerifyAttestationResponse{},
		shared.Back{}).
	State(shared.StateProfileAccountDeleted).
	InitialState(shared.StatePreflight, shared.StateProfileInit).
	ErrorState(shared.StateError).
	BeforeEachAction(profile.RefreshSessionUser{}).
	BeforeState(shared.StateProfileInit, profile.GetProfileData{}).
	AfterState(shared.StateProfileWebauthnCredentialVerification, shared.WebauthnCredentialSave{}).
	AfterState(shared.StatePasscodeConfirmation, shared.EmailPersistVerifiedStatus{}).
	SubFlows(
		CapabilitiesSubFlow,
		CredentialUsageSubFlow).
	TTL(24 * time.Hour).
	Debug(true)
