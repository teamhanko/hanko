package flow_api_test

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SubmitEmail struct{}

func (m SubmitEmail) GetName() flowpilot.MethodName {
	return MethodSubmitEmail
}

func (m SubmitEmail) GetDescription() string {
	return "Enter an email address to sign in or sign up."
}

func (m SubmitEmail) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.EmailInput("email").Required(true).Preserve(true))
}

func (m SubmitEmail) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.FormDataInvalidError)
	}

	_ = c.CopyInputValuesToStash("email")

	user, _ := models.MyUsers.FindByEmail(c.Input().Get("email").String())

	if user != nil {
		if myFlowConfig.isEnabled(FlowOptionPasswords) {
			return c.ContinueFlow(StateLoginWithPassword)
		}

		if !myFlowConfig.isEnabled(FlowOptionSecondFactorFlow) {
			initPasscode(c, c.Stash().Get("email").String(), false)
			return c.ContinueFlow(StateLoginWithPasscode)
		}

		return c.ContinueFlowWithError(StateError, flowpilot.FlowDiscontinuityError)
	}

	return c.ContinueFlow(StateConfirmAccountCreation)
}

type GetWAChallenge struct{}

func (m GetWAChallenge) GetName() flowpilot.MethodName {
	return MethodGetWAChallenge
}

func (m GetWAChallenge) GetDescription() string {
	return "Get the passkey challenge."
}

func (m GetWAChallenge) Initialize(_ flowpilot.InitializationContext) {}

func (m GetWAChallenge) Execute(c flowpilot.ExecutionContext) error {
	initPasskey(c)
	return c.ContinueFlow(StateLoginWithPasskey)
}

type VerifyWAPublicKey struct{}

func (m VerifyWAPublicKey) GetName() flowpilot.MethodName {
	return MethodVerifyWAPublicKey
}

func (m VerifyWAPublicKey) GetDescription() string {
	return "Verifies the challenge."
}

func (m VerifyWAPublicKey) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("passkey_public_key").Required(true).CompareWithStash(true))
}

func (m VerifyWAPublicKey) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.FormDataInvalidError)
	}

	return c.ContinueFlow(StateSuccess)
}

type SubmitExistingPassword struct{}

func (m SubmitExistingPassword) GetName() flowpilot.MethodName {
	return MethodSubmitExistingPassword
}

func (m SubmitExistingPassword) GetDescription() string {
	return "Enter your password to sign in."
}

func (m SubmitExistingPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("password").Required(true))
}

func (m SubmitExistingPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.FormDataInvalidError)
	}

	email := c.Stash().Get("email").String()
	user, _ := models.MyUsers.FindByEmail(email)

	if user != nil && user.Password == c.Input().Get("password").String() {
		if myFlowConfig.isEnabled(FlowOptionSecondFactorFlow) && user.Passcode2faEnabled {
			initPasscode(c, email, true)
			return c.ContinueFlow(StateLoginWithPasscode2FA)
		}

		if user.PasskeySynced {
			return c.ContinueFlow(StateSuccess)
		}

		return c.ContinueFlow(StateConfirmPasskeyCreation)
	}

	c.Input().SetError("password", flowpilot.ValueInvalidError)

	return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.FormDataInvalidError)
}

type RequestRecovery struct{}

func (m RequestRecovery) GetName() flowpilot.MethodName {
	return MethodRequestRecovery
}

func (m RequestRecovery) GetDescription() string {
	return "Request passcode recovery to set a new password."
}

func (m RequestRecovery) Initialize(c flowpilot.InitializationContext) {
	if myFlowConfig.isEnabled(FlowOptionSecondFactorFlow) {
		c.SuspendMethod()
	}
}

func (m RequestRecovery) Execute(c flowpilot.ExecutionContext) error {
	initPasscode(c, c.Stash().Get("email").String(), false)
	return c.ContinueFlow(StateRecoverPasswordViaPasscode)
}

type SubmitPasscodeCode struct{}

func (m SubmitPasscodeCode) GetName() flowpilot.MethodName {
	return MethodSubmitPasscodeCode
}

func (m SubmitPasscodeCode) GetDescription() string {
	return "Enter the passcode sent via email."
}

func (m SubmitPasscodeCode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(
		flowpilot.StringInput("passcode_id").Required(true).Hidden(true).Preserve(true).CompareWithStash(true),
		flowpilot.StringInput("code").Required(true).MinLength(6).MaxLength(6).CompareWithStash(true),
		flowpilot.StringInput("passcode_2fa_token").Required(true).Hidden(true).Preserve(true).CompareWithStash(true).ConditionalIncludeOnState(StateLoginWithPasscode2FA),
	)
}

func (m SubmitPasscodeCode) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.FormDataInvalidError)
	}

	if c.CurrentStateEquals(StateRecoverPasswordViaPasscode) {
		return c.ContinueFlow(StateUpdateExistingPassword)
	}

	user, _ := models.MyUsers.FindByEmail(c.Stash().Get("email").String())

	if c.CurrentStateEquals(StateLoginWithPasscode, StateVerifyEmailViaPasscode, StateLoginWithPasscode2FA) {
		if user != nil && user.PasskeySynced {
			return c.ContinueFlow(StateSuccess)
		}

		return c.ContinueFlow(StateConfirmPasskeyCreation)
	}

	return c.ContinueFlow(StateSuccess)
}

type CreateUser struct{}

func (m CreateUser) GetName() flowpilot.MethodName {
	return MethodCreateUser
}

func (m CreateUser) GetDescription() string {
	return "Confirm account creation."
}

func (m CreateUser) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs()
}

func (m CreateUser) Execute(c flowpilot.ExecutionContext) error {
	if myFlowConfig.isEnabled(FlowOptionPasswords) {
		return c.ContinueFlow(StatePasswordCreation)
	}

	if myFlowConfig.isEnabled(FlowOptionEmailVerification) {
		initPasscode(c, c.Stash().Get("email").String(), false)
		return c.ContinueFlow(StateLoginWithPasscode)
	}

	return c.ContinueFlow(StateConfirmPasskeyCreation)
}

type SubmitNewPassword struct{}

func (m SubmitNewPassword) GetName() flowpilot.MethodName {
	return MethodSubmitNewPassword
}

func (m SubmitNewPassword) GetDescription() string {
	return "Enter a new password"
}

func (m SubmitNewPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("password").Required(true).MinLength(8).MaxLength(32))
}

func (m SubmitNewPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.FormDataInvalidError)
	}

	if c.CurrentStateEquals(StateUpdateExistingPassword) {
		return c.ContinueFlow(StateSuccess)
	}

	if myFlowConfig.isEnabled(FlowOptionEmailVerification) {
		initPasscode(c, c.Stash().Get("email").String(), false)
		return c.ContinueFlow(StateVerifyEmailViaPasscode)
	}

	return c.ContinueFlow(StateConfirmPasskeyCreation)

}

type GetWAAssertion struct{}

func (m GetWAAssertion) GetName() flowpilot.MethodName {
	return MethodGetWAAssertion
}

func (m GetWAAssertion) GetDescription() string {
	return "Creates a new passkey."
}

func (m GetWAAssertion) Initialize(_ flowpilot.InitializationContext) {}

func (m GetWAAssertion) Execute(c flowpilot.ExecutionContext) error {
	initPasskey(c)
	return c.ContinueFlow(StateCreatePasskey)
}

type VerifyWAAssertion struct{}

func (m VerifyWAAssertion) GetName() flowpilot.MethodName {
	return MethodVerifyWAAssertion
}

func (m VerifyWAAssertion) GetDescription() string {
	return "Verifies the passkey creation."
}

func (m VerifyWAAssertion) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("passkey_public_key").Required(true).CompareWithStash(true))
}

func (m VerifyWAAssertion) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.FormDataInvalidError)
	}

	return c.ContinueFlow(StateSuccess)
}

type SkipPasskeyCreation struct{}

func (m SkipPasskeyCreation) GetName() flowpilot.MethodName {
	return MethodSkipPasskeyCreation
}

func (m SkipPasskeyCreation) GetDescription() string {
	return "Skips the onboarding process."
}

func (m SkipPasskeyCreation) Initialize(_ flowpilot.InitializationContext) {}

func (m SkipPasskeyCreation) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StateSuccess)
}

type Back struct{}

func (m Back) GetName() flowpilot.MethodName {
	return MethodBack
}

func (m Back) GetDescription() string {
	return "Go one step back."
}

func (m Back) Initialize(_ flowpilot.InitializationContext) {}

func (m Back) Execute(c flowpilot.ExecutionContext) error {
	if previousState := c.GetPreviousState(); previousState != nil {
		return c.ContinueFlow(*previousState)
	}

	return c.ContinueFlow(c.GetInitialState())
}
