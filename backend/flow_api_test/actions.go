package flow_api_test

import (
	"errors"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type ContinueToFinal struct{}

func (m ContinueToFinal) GetName() flowpilot.ActionName {
	return ActionContinueToFinal
}

func (m ContinueToFinal) GetDescription() string {
	return ""
}

func (m ContinueToFinal) Initialize(c flowpilot.InitializationContext) {}

func (m ContinueToFinal) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StateSecondSubFlowFinal)
}

type EndSubFlow struct{}

func (m EndSubFlow) GetName() flowpilot.ActionName {
	return ActionEndSubFlow
}

func (m EndSubFlow) GetDescription() string {
	return ""
}

func (m EndSubFlow) Initialize(c flowpilot.InitializationContext) {}

func (m EndSubFlow) Execute(c flowpilot.ExecutionContext) error {
	return c.EndSubFlow()
}

type StartSecondSubFlow struct{}

func (m StartSecondSubFlow) GetName() flowpilot.ActionName {
	return ActionStartSecondSubFlow
}

func (m StartSecondSubFlow) GetDescription() string {
	return ""
}

func (m StartSecondSubFlow) Initialize(c flowpilot.InitializationContext) {}

func (m StartSecondSubFlow) Execute(c flowpilot.ExecutionContext) error {
	return c.StartSubFlow(StateSecondSubFlowInit)
}

type StartFirstSubFlow struct{}

func (m StartFirstSubFlow) GetName() flowpilot.ActionName {
	return ActionStartFirstSubFlow
}

func (m StartFirstSubFlow) GetDescription() string {
	return ""
}

func (m StartFirstSubFlow) Initialize(c flowpilot.InitializationContext) {}

func (m StartFirstSubFlow) Execute(c flowpilot.ExecutionContext) error {
	return c.StartSubFlow(StateFirstSubFlowInit, StateThirdSubFlowInit, StateSuccess)
}

type SubmitEmail struct{}

func (m SubmitEmail) GetName() flowpilot.ActionName {
	return ActionSubmitEmail
}

func (m SubmitEmail) GetDescription() string {
	return "Enter an email address to sign in or sign up."
}

func (m SubmitEmail) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.EmailInput("email").Required(true).Preserve(true))
}

func (m SubmitEmail) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
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

		return c.ContinueFlowWithError(StateError, flowpilot.ErrorFlowDiscontinuity)
	}

	return c.ContinueFlow(StateConfirmAccountCreation)
}

type GetWAChallenge struct{}

func (m GetWAChallenge) GetName() flowpilot.ActionName {
	return ActionGetWAChallenge
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

func (m VerifyWAPublicKey) GetName() flowpilot.ActionName {
	return ActionVerifyWAPublicKey
}

func (m VerifyWAPublicKey) GetDescription() string {
	return "Verifies the challenge."
}

func (m VerifyWAPublicKey) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("passkey_public_key").Required(true).CompareWithStash(true))
}

func (m VerifyWAPublicKey) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	return c.ContinueFlow(StateSuccess)
}

type SubmitExistingPassword struct{}

func (m SubmitExistingPassword) GetName() flowpilot.ActionName {
	return ActionSubmitExistingPassword
}

func (m SubmitExistingPassword) GetDescription() string {
	return "Enter your password to sign in."
}

func (m SubmitExistingPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("password").Required(true))
}

func (m SubmitExistingPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
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

	c.Input().SetError("password", flowpilot.ErrorValueInvalid.Wrap(errors.New("password does not match")))

	return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
}

type RequestRecovery struct{}

func (m RequestRecovery) GetName() flowpilot.ActionName {
	return ActionRequestRecovery
}

func (m RequestRecovery) GetDescription() string {
	return "Request passcode recovery to set a new password."
}

func (m RequestRecovery) Initialize(c flowpilot.InitializationContext) {
	if myFlowConfig.isEnabled(FlowOptionSecondFactorFlow) {
		c.SuspendAction()
	}
}

func (m RequestRecovery) Execute(c flowpilot.ExecutionContext) error {
	initPasscode(c, c.Stash().Get("email").String(), false)
	return c.ContinueFlow(StateRecoverPasswordViaPasscode)
}

type SubmitPasscodeCode struct{}

func (m SubmitPasscodeCode) GetName() flowpilot.ActionName {
	return ActionSubmitPasscodeCode
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
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
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

func (m CreateUser) GetName() flowpilot.ActionName {
	return ActionCreateUser
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

func (m SubmitNewPassword) GetName() flowpilot.ActionName {
	return ActionSubmitNewPassword
}

func (m SubmitNewPassword) GetDescription() string {
	return "Enter a new password"
}

func (m SubmitNewPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("password").Required(true).MinLength(8).MaxLength(32))
}

func (m SubmitNewPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
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

func (m GetWAAssertion) GetName() flowpilot.ActionName {
	return ActionGetWAAssertion
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

func (m VerifyWAAssertion) GetName() flowpilot.ActionName {
	return ActionVerifyWAAssertion
}

func (m VerifyWAAssertion) GetDescription() string {
	return "Verifies the passkey creation."
}

func (m VerifyWAAssertion) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("passkey_public_key").Required(true).CompareWithStash(true))
}

func (m VerifyWAAssertion) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	return c.ContinueFlow(StateSuccess)
}

type SkipPasskeyCreation struct{}

func (m SkipPasskeyCreation) GetName() flowpilot.ActionName {
	return ActionSkipPasskeyCreation
}

func (m SkipPasskeyCreation) GetDescription() string {
	return "Skips the onboarding process."
}

func (m SkipPasskeyCreation) Initialize(_ flowpilot.InitializationContext) {}

func (m SkipPasskeyCreation) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StateSuccess)
}

type Back struct{}

func (m Back) GetName() flowpilot.ActionName {
	return ActionBack
}

func (m Back) GetDescription() string {
	return "Go one step back."
}

func (m Back) Initialize(_ flowpilot.InitializationContext) {}

func (m Back) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueToPreviousState()
}

type BeforeStateAction struct{}

func (m BeforeStateAction) Execute(c flowpilot.HookExecutionContext) error {
	return c.Payload().Set("before_action_executed", true)
}
