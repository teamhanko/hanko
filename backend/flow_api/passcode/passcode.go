package passcode

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/passcode/actions"
	"github.com/teamhanko/hanko/backend/flow_api/passcode/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared/hooks"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
)

func NewPasscodeSubFlow(cfg config.Config, persister persistence.Persister, passcodeService services.Passcode, httpContext echo.Context) (flowpilot.SubFlow, error) {
	return flowpilot.NewSubFlow().
		State(states.StatePasscodeConfirmation, actions.NewSubmitPasscode(cfg, persister)).
		BeforeState(states.StatePasscodeConfirmation, hooks.NewSendPasscode(passcodeService, httpContext)).
		Build()
}
