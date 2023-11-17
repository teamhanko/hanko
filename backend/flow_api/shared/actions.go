package shared

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
)

const (
	ActionSendCapabilities                       flowpilot.ActionName = "send_capabilities"
	ActionContinueToPasscodeConfirmation         flowpilot.ActionName = "continue_to_passcode_confirmation"
	ActionContinueToPasscodeConfirmationRecovery flowpilot.ActionName = "continue_to_passcode_confirmation_recovery"
	ActionContinueToLoginMethodChooser           flowpilot.ActionName = "continue_to_login_method_chooser"
	ActionLoginWithOauth                         flowpilot.ActionName = "login_with_oauth"
	ActionLoginWithPassword                      flowpilot.ActionName = "login_with_password"
	ActionSubmitRegistrationIdentifier           flowpilot.ActionName = "submit_registration_identifier"
	ActionSubmitLoginIdentifier                  flowpilot.ActionName = "submit_login_identifier"
	ActionSubmitPasscode                         flowpilot.ActionName = "submit_email_passcode"
	ActionGetWARequestOptions                    flowpilot.ActionName = "get_wa_request_options"
	ActionSendWAAssertionResponse                flowpilot.ActionName = "send_wa_request_response"
	ActionGetWACreationOptions                   flowpilot.ActionName = "get_wa_creation_options"
	ActionSendWAAttestationResponse              flowpilot.ActionName = "send_wa_attestation_response"
	ActionSubmitPassword                         flowpilot.ActionName = "submit_password"
	ActionSubmitNewPassword                      flowpilot.ActionName = "submit_new_password"
	ActionSubmitTOTPCode                         flowpilot.ActionName = "submit_totp_code"
	ActionGenerateRecoveryCodes                  flowpilot.ActionName = "generate_recovery_codes"
	ActionStart2FARecovery                       flowpilot.ActionName = "start_2fa_recovery"
	ActionSubmitRecoveryCode                     flowpilot.ActionName = "submit_recovery_code"

	ActionSwitch   flowpilot.ActionName = "switch"
	ActionBack     flowpilot.ActionName = "back"
	ActionSkip     flowpilot.ActionName = "skip"
	ActionContinue flowpilot.ActionName = "continue"
)

type Dependencies struct {
	Cfg         config.Config
	Tx          *pop.Connection
	Persister   persistence.Persister
	HttpContext echo.Context
}

type Action struct{}

func (a *Action) GetDepsForExecution(c flowpilot.ExecutionContext) *Dependencies {
	return c.Get("dependencies").(*Dependencies)
}

func (a *Action) GetDepsForInitialization(c flowpilot.InitializationContext) *Dependencies {
	return c.Get("dependencies").(*Dependencies)
}
