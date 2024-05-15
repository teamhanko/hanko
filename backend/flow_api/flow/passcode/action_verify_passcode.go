package passcode

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/constants"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type VerifyPasscode struct {
	shared.Action
}

func (a VerifyPasscode) GetName() flowpilot.ActionName {
	return constants.ActionVerifyPasscode
}

func (a VerifyPasscode) GetDescription() string {
	return "Enter a passcode."
}

func (a VerifyPasscode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("code").Required(true))
}

func (a VerifyPasscode) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if !c.Stash().Get("passcode_id").Exists() {
		return errors.New("passcode_id does not exist in the stash")
	}

	passcodeID := uuid.FromStringOrNil(c.Stash().Get("passcode_id").String())
	err := deps.PasscodeService.VerifyPasscodeCode(deps.Tx, passcodeID, c.Input().Get("code").String())
	if err != nil {
		if errors.Is(err, services.ErrorPasscodeInvalid) ||
			errors.Is(err, services.ErrorPasscodeNotFound) ||
			errors.Is(err, services.ErrorPasscodeExpired) {

			if c.Stash().Get("login_method").Exists() {
				err = deps.AuditLogger.CreateWithConnection(
					deps.Tx,
					deps.HttpContext,
					models.AuditLogLoginFailure,
					&models.User{ID: uuid.FromStringOrNil(c.Stash().Get("user_id").String())},
					err,
					auditlog.Detail("login_method", "passcode"),
					auditlog.Detail("flow_id", c.GetFlowID()))

				if err != nil {
					return fmt.Errorf("could not create audit log: %w", err)
				}
			}

			return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorPasscodeInvalid)
		}

		if errors.Is(err, services.ErrorPasscodeMaxAttemptsReached) {
			if c.Stash().Get("login_method").Exists() {
				err = deps.AuditLogger.CreateWithConnection(
					deps.Tx,
					deps.HttpContext,
					models.AuditLogLoginFailure,
					&models.User{ID: uuid.FromStringOrNil(c.Stash().Get("user_id").String())},
					err,
					auditlog.Detail("login_method", "passcode"),
					auditlog.Detail("flow_id", c.GetFlowID()))

				if err != nil {
					return fmt.Errorf("could not create audit log: %w", err)
				}
			}

			return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorPasscodeMaxAttemptsReached)
		}

		return fmt.Errorf("failed to verify passcode: %w", err)
	}

	err = c.Stash().Delete("passcode_id")
	if err != nil {
		return fmt.Errorf("failed to delete passcode_id from stash: %w", err)
	}

	err = c.Stash().Delete("passcode_email")
	if err != nil {
		return fmt.Errorf("failed to delete passcode_email from stash: %w", err)
	}

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("account does not exist")))
	}

	err = c.Stash().Set("email_verified", true) // TODO: maybe change attribute path
	if err != nil {
		return err
	}

	if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		err = c.Stash().Set("allow_skip_onboarding", true)
		if err != nil {
			return fmt.Errorf("failed to set allow_skip_onboarding to stash: %w", err)
		}
	}

	return c.EndSubFlow()
}

func (a VerifyPasscode) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
