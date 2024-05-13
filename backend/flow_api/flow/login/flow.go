package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login_method_chooser"
	"github.com/teamhanko/hanko/backend/flow_api/flow/login_password"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

const (
	StateLoginInit    flowpilot.StateName = "login_init"
	StateLoginPasskey flowpilot.StateName = "login_passkey"
)

const (
	ActionWebauthnGenerateRequestOptions  flowpilot.ActionName = "webauthn_generate_request_options"
	ActionWebauthnVerifyAssertionResponse flowpilot.ActionName = "webauthn_verify_assertion_response"
	ActionContinueWithLoginIdentifier     flowpilot.ActionName = "continue_with_login_identifier"
)

var Flow = flowpilot.NewFlow("/login").
	State(StateLoginInit,
		ContinueWithLoginIdentifier{},
		WebauthnGenerateRequestOptions{},
		WebauthnVerifyAssertionResponse{},
		shared.ThirdPartyOAuth{}).
	BeforeState(StateLoginInit, WebauthnGenerateRequestOptionsForConditionalUi{}).
	State(shared.StateThirdPartyOAuth, shared.ExchangeToken{}).
	State(StateLoginPasskey, WebauthnVerifyAssertionResponse{}, shared.Back{}).
	BeforeState(shared.StateSuccess, shared.IssueSession{}).
	State(shared.StateSuccess).
	State(shared.StateError).
	SubFlows(capabilities.SubFlow, passkey_onboarding.SubFlow, passcode.SubFlow, login_method_chooser.SubFlow, login_password.SubFlow).
	AfterState(passkey_onboarding.StateOnboardingVerifyPasskeyAttestation, shared.WebauthnCredentialSave{}).
	InitialState(capabilities.StatePreflight, StateLoginInit).
	BeforeState(passcode.StatePasscodeConfirmation, SelectPasscodeTemplate{}).
	AfterState(passcode.StatePasscodeConfirmation, shared.EmailPersistVerifiedStatus{}).
	ErrorState(shared.StateError).
	TTL(10 * time.Minute)
