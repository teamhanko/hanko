package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type EmailDelete struct {
	shared.Action
}

func (a EmailDelete) GetName() flowpilot.ActionName {
	return ActionEmailDelete
}

func (a EmailDelete) GetDescription() string {
	return "Delete an email address."
}

func (a EmailDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Identifier.Email.Enabled {
		c.SuspendAction()
	} else {
		c.AddInputs(flowpilot.StringInput("email_id").Required(true).Hidden(true))
	}
}

func (a EmailDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	emailToBeDeletedId := uuid.FromStringOrNil(c.Input().Get("email_id").String())
	emailToBeDeletedModel := userModel.GetEmailById(emailToBeDeletedId)
	if emailToBeDeletedModel == nil {
		return c.ContinueFlowWithError(
			c.GetCurrentState(),
			flowpilot.ErrorFormDataInvalid.Wrap(errors.New("unknown email")),
		)
	}

	if emailToBeDeletedModel.IsPrimary() {
		if !deps.Cfg.Identifier.Email.Optional {
			return c.ContinueFlowWithError(
				c.GetCurrentState(),
				flowpilot.ErrorOperationNotPermitted.Wrap(errors.New("cannot delete primary email")),
			)
		} else {
			err := deps.Persister.GetPrimaryEmailPersisterWithConnection(deps.Tx).Delete(*emailToBeDeletedModel.PrimaryEmail)
			if err != nil {
				return fmt.Errorf("could not delete primary email: %w", err)
			}

			err = deps.Persister.GetEmailPersisterWithConnection(deps.Tx).Delete(*emailToBeDeletedModel)
			if err != nil {
				return fmt.Errorf("could not delete email: %w", err)
			}
		}
	}

	return c.ContinueFlow(StateProfileInit)
}
