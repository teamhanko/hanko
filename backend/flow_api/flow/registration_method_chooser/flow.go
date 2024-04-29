package registration_method_chooser

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/registration_register_password"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StateRegistrationMethodChooser flowpilot.StateName = "registration_method_chooser"
)

const (
	ActionContinueToPasswordRegistration flowpilot.ActionName = "continue_to_password_registration"
	ActionContinueToPasskeyRegistration  flowpilot.ActionName = "continue_to_passkey_registration"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StateRegistrationMethodChooser,
		ContinueToPasskeyCreation{},
		ContinueToPasswordRegistration{},
		shared.Back{},
		shared.Skip{}).
	SubFlows(passkey_onboarding.SubFlow, registration_register_password.SubFlow).
	MustBuild()
