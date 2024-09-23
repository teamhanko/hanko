package flow

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/credential_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/credential_usage"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login"
	"github.com/teamhanko/hanko/backend/flow_api/flow/mfa_creation"
	"github.com/teamhanko/hanko/backend/flow_api/flow/mfa_usage"
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
	State(shared.StateLoginInit,
		credential_usage.ContinueWithLoginIdentifier{},
		credential_usage.WebauthnGenerateRequestOptions{},
		credential_usage.WebauthnVerifyAssertionResponse{},
		shared.ThirdPartyOAuth{}).
	State(shared.StateLoginPasskey,
		credential_usage.WebauthnVerifyAssertionResponse{},
		shared.Back{}).
	State(shared.StateThirdParty,
		shared.ExchangeToken{}).
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

var MFACreationSubFlow = flowpilot.NewSubFlow(shared.FlowMFACreation).
	State(shared.StateMFAMethodChooser,
		mfa_creation.ContinueToOTPSecretCreation{},
		mfa_creation.ContinueToSecurityKeyCreation{},
		mfa_creation.SkipMFA{}).
	BeforeState(shared.StateMFAOTPSecretCreation,
		mfa_creation.OTPSecretGenerate{}).
	State(shared.StateMFAOTPSecretCreation,
		mfa_creation.OTPCodeVerify{},
		shared.Back{}).
	State(shared.StateMFASecurityKeyCreation,
		mfa_creation.WebauthnGenerateCreationOptionsForSecurityKeys{},
		shared.Back{}).
	MustBuild()

var MFAUsageSubFlow = flowpilot.NewSubFlow(shared.FlowMFAUsage).
	State(shared.StateLoginSecurityKey,
		mfa_usage.WebauthnGenerateRequestOptionsSecurityKey{},
		mfa_usage.ContinueToLoginOTP{}).
	State(shared.StateLoginOTP,
		mfa_usage.OTPCodeValidate{},
		mfa_usage.ContinueToLoginSecurityKey{}).
	MustBuild()

func NewLoginFlow(debug bool) flowpilot.Flow {
	return flowpilot.NewFlow(shared.FlowLogin).
		State(shared.StateSuccess).
		InitialState(shared.StatePreflight, shared.StateLoginInit).
		ErrorState(shared.StateError).
		BeforeState(shared.StateLoginInit,
			login.WebauthnGenerateRequestOptionsForConditionalUi{}).
		BeforeState(shared.StateSuccess,
			shared.IssueSession{},
			shared.GetUserData{}).
		AfterState(shared.StateOnboardingVerifyPasskeyAttestation,
			shared.WebauthnCredentialSave{}).
		AfterState(shared.StatePasscodeConfirmation,
			shared.EmailPersistVerifiedStatus{}).
		AfterState(shared.StatePasswordCreation,
			shared.PasswordSave{}).
		AfterState(shared.StateOnboardingEmail, login.CreateEmail{}).
		AfterState(shared.StatePasscodeConfirmation, login.CreateEmail{}).
		AfterFlow(shared.FlowCredentialUsage, login.ScheduleOnboardingStates{}).
		AfterFlow(shared.FlowUserDetails, registration.ScheduleMFACreationStates{}).
		SubFlows(
			CapabilitiesSubFlow,
			CredentialUsageSubFlow,
			CredentialOnboardingSubFlow,
			UserDetailsSubFlow,
			MFACreationSubFlow,
			MFAUsageSubFlow).
		TTL(24 * time.Hour).
		Debug(debug).
		MustBuild()
}

func NewRegistrationFlow(debug bool) flowpilot.Flow {
	return flowpilot.NewFlow(shared.FlowRegistration).
		State(shared.StateRegistrationInit,
			registration.RegisterLoginIdentifier{},
			shared.ThirdPartyOAuth{}).
		State(shared.StateThirdParty,
			shared.ExchangeToken{}).
		State(shared.StateSuccess).
		InitialState(shared.StatePreflight,
			shared.StateRegistrationInit).
		ErrorState(shared.StateError).
		BeforeState(shared.StateSuccess,
			shared.IssueSession{},
			shared.GetUserData{},
			registration.CreateUser{}).
		SubFlows(
			CapabilitiesSubFlow,
			CredentialUsageSubFlow,
			CredentialOnboardingSubFlow,
			UserDetailsSubFlow,
			MFACreationSubFlow).
		TTL(24 * time.Hour).
		Debug(debug).
		MustBuild()
}

func NewProfileFlow(debug bool) flowpilot.Flow {
	return flowpilot.NewFlow(shared.FlowProfile).
		State(shared.StateProfileInit,
			profile.AccountDelete{},
			profile.ContinueToOTPSecretCreation{},
			profile.EmailCreate{},
			profile.EmailDelete{},
			profile.EmailSetPrimary{},
			profile.EmailVerify{},
			profile.OTPSecretDelete{},
			profile.PasswordCreate{},
			profile.PasswordDelete{},
			profile.PasswordUpdate{},
			profile.SecurityKeyCreate{},
			profile.SecurityKeyDelete{},
			profile.UsernameCreate{},
			profile.UsernameDelete{},
			profile.UsernameUpdate{},
			profile.WebauthnCredentialCreate{},
			profile.WebauthnCredentialDelete{},
			profile.WebauthnCredentialRename{},
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
			CredentialOnboardingSubFlow,
			CredentialUsageSubFlow,
			MFACreationSubFlow).
		TTL(24 * time.Hour).
		Debug(debug).
		MustBuild()
}
