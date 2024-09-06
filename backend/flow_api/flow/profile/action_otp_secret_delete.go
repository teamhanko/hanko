package profile

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type OTPSecretDelete struct {
	shared.Action
}

func (a OTPSecretDelete) GetName() flowpilot.ActionName {
	return shared.ActionOTPSecretDelete
}

func (a OTPSecretDelete) GetDescription() string {
	return "Delete an OTP secret"
}

func (a OTPSecretDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if userModel.OTPSecret == nil {
		c.SuspendAction()
		return
	}

	if deps.Cfg.MFA.Enabled && !deps.Cfg.MFA.Optional && len(userModel.GetSecurityKeys()) > 0 {
		c.SuspendAction()
		return
	}
}

func (a OTPSecretDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	err := deps.Persister.GetOTPSecretPersisterWithConnection(deps.Tx).Delete(userModel.OTPSecret)
	if err != nil {
		return fmt.Errorf("could not delete otp secret: %w", err)
	}

	userModel.DeleteOTPSecret()

	return c.Continue(shared.StateProfileInit)
}
