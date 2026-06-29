package mfa_creation

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v3/flow_api/services"
	"github.com/teamhanko/hanko/backend/v3/flowpilot"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
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

	matchedStep, err := deps.TOTPService.ValidateCode(code, &models.OTPSecret{Secret: secret}, time.Now().UTC())
	if err != nil {
		if errors.Is(err, services.ErrTOTPCodeInvalid) {
			c.Input().SetError("otp_code", shared.ErrorPasscodeInvalid)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}
		return fmt.Errorf("totp validation error: %w", err)
	}

	// Stash the matched step so hook_create_user can set it on the OTPSecret during registration.
	if err := c.Stash().Set(shared.StashPathOTPLastValidatedStep, matchedStep); err != nil {
		return fmt.Errorf("failed to stash otp last validated step: %w", err)
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

		otpSecretModel := models.NewOTPSecret(userID, secret, deps.TenantID)
		// Set LastValidatedStep to the step just consumed during setup verification.
		// This prevents the setup code from being replayed as the first MFA login
		// code (RFC 6238 §5.2): the persisted secret already has that step consumed,
		// so the replay check in TOTPService.ValidateCode will reject it.
		otpSecretModel.LastValidatedStep = nulls.NewInt64(matchedStep)

		err := deps.Persister.GetOTPSecretPersisterWithConnection(deps.Tx).Create(*otpSecretModel)
		if err != nil {
			return fmt.Errorf("could not create OTP secret: %w", err)
		}

		if userModel != nil {
			// Send user an email informing of new MFA method
			if deps.Cfg.SecurityNotifications.Notifications.MFACreate.Enabled {
				deps.SecurityNotificationService.SendNotification(deps.Tx, services.SendSecurityNotificationParams{
					TenantID:     deps.TenantID,
					EmailAddress: userModel.Emails.GetPrimary().Address,
					Template:     "mfa_create",
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
			deps.TenantID,
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
