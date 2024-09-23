package credential_usage

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/registration"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type VerifyPasscode struct {
	shared.Action
}

func (a VerifyPasscode) GetName() flowpilot.ActionName {
	return shared.ActionVerifyPasscode
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
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	if !c.Stash().Get(shared.StashPathPasscodeID).Exists() {
		return errors.New("passcode_id does not exist in the stash")
	}

	passcodeID := uuid.FromStringOrNil(c.Stash().Get(shared.StashPathPasscodeID).String())
	err := deps.PasscodeService.VerifyPasscodeCode(deps.Tx, passcodeID, c.Input().Get("code").String())
	if err != nil {
		if errors.Is(err, services.ErrorPasscodeInvalid) ||
			errors.Is(err, services.ErrorPasscodeNotFound) ||
			errors.Is(err, services.ErrorPasscodeExpired) {

			if c.Stash().Get(shared.StashPathLoginMethod).Exists() {
				err = deps.AuditLogger.CreateWithConnection(
					deps.Tx,
					deps.HttpContext,
					models.AuditLogLoginFailure,
					&models.User{ID: uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())},
					err,
					auditlog.Detail("login_method", "passcode"),
					auditlog.Detail("flow_id", c.GetFlowID()))

				if err != nil {
					return fmt.Errorf("could not create audit log: %w", err)
				}
			}

			return c.Error(shared.ErrorPasscodeInvalid)
		}

		if errors.Is(err, services.ErrorPasscodeMaxAttemptsReached) {
			if c.Stash().Get(shared.StashPathLoginMethod).Exists() {
				err = deps.AuditLogger.CreateWithConnection(
					deps.Tx,
					deps.HttpContext,
					models.AuditLogLoginFailure,
					&models.User{ID: uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())},
					err,
					auditlog.Detail("login_method", "passcode"),
					auditlog.Detail("flow_id", c.GetFlowID()))

				if err != nil {
					return fmt.Errorf("could not create audit log: %w", err)
				}
			}

			return c.Error(shared.ErrorPasscodeMaxAttemptsReached)
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

	if !c.Stash().Get(shared.StashPathUserID).Exists() {
		return c.Error(flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("account does not exist")))
	}

	err = c.Stash().Set(shared.StashPathEmailVerified, true)
	if err != nil {
		return err
	}

	err = c.Stash().Set(shared.StashPathUserHasEmails, true)
	if err != nil {
		return err
	}

	c.PreventRevert()

	passwordCreationScheduled := false

	for _, state := range c.GetScheduledStates() {
		if state == shared.StatePasswordCreation {
			passwordCreationScheduled = true
			break
		}
	}

	if c.GetFlowName() == shared.FlowRegistration && !passwordCreationScheduled {
		if err = c.ExecuteHook(registration.ScheduleMFACreationStates{}); err != nil {
			return err
		}
	}

	return c.Continue()
}
