package flows

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/actions"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/hooks"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
)

func NewPasscodeSubFlow(cfg config.Config, persister persistence.Persister, passcodeService services.Passcode, httpContext echo.Context) (flowpilot.SubFlow, error) {
	return flowpilot.NewSubFlow().
		State(common.StatePasscodeConfirmation, actions.NewSubmitPasscode(cfg, persister)).
		BeforeState(common.StatePasscodeConfirmation, hooks.NewSendPasscode(passcodeService, httpContext)).
		Build()
}
