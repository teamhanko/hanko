package mfa_creation

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/pquerna/otp/totp"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type OTPCodeVerify struct {
	shared.Action
}

func (a OTPCodeVerify) GetName() flowpilot.ActionName {
	return shared.ActionOTPCodeVerify
}

func (a OTPCodeVerify) GetDescription() string {
	return "Verify an OTP code"
}

func (a OTPCodeVerify) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.StringInput("otp_code").Required(true))

	if !deps.Cfg.MFA.TOTP.Enabled {
		c.SuspendAction()
	}
}

func (a OTPCodeVerify) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	code := c.Input().Get("otp_code").String()
	secret := c.Stash().Get(shared.StashPathOTPSecret).String()

	if !totp.Validate(code, secret) {
		c.Input().SetError("otp_code", shared.ErrorPasscodeInvalid)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	_ = c.Stash().Set(shared.StashPathUserHasOTPSecret, true)

	if c.GetFlowName() != shared.FlowRegistration {
		var userID uuid.UUID
		var userModel *models.User
		if c.GetFlowName() == shared.FlowLogin {
			userID = uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())
		} else if c.GetFlowName() == shared.FlowProfile {
			user, ok := c.Get("session_user").(*models.User)
			if !ok {
				return c.Error(flowpilot.ErrorOperationNotPermitted)
			}
			userModel = user
			userID = userModel.ID
		}

		otpSecretModel := models.NewOTPSecret(userID, secret)

		err := deps.Persister.GetOTPSecretPersisterWithConnection(deps.Tx).Create(*otpSecretModel)
		if err != nil {
			return fmt.Errorf("could not create OTP secret: %w", err)
		}

		if userModel != nil {
			// Send MFA enabled notification if this is the first MFA method
			if !userModel.HasMFAEnabled() && deps.Cfg.SecurityNotifications.Notifications.MFAEnabled.Enabled {
				deps.SecurityNotificationService.SendNotification(deps.Tx, services.SendSecurityNotificationParams{
					EmailAddress: userModel.Emails.GetPrimary().Address,
					Template:     "mfa_enabled",
					HttpContext:  deps.HttpContext,
					UserContext:  *userModel,
				})
			}
		}

		err = deps.AuditLogger.CreateWithConnection(
			deps.Tx,
			deps.HttpContext,
			models.AuditLogOTPCreated,
			&models.User{ID: userID},
			nil,
			auditlog.Detail("otp_secret", otpSecretModel.ID),
			auditlog.Detail("flow_id", c.GetFlowID()),
		)

		if err != nil {
			return fmt.Errorf("could not create audit log: %w", err)
		}
	}

	c.PreventRevert()

	return c.Continue()
}
