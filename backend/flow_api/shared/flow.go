package shared

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
)

const (
	StateSuccess flowpilot.StateName = "success"
	StateError   flowpilot.StateName = "error"
)

const (
	ActionBack flowpilot.ActionName = "back"
)

type Dependencies struct {
	Cfg             config.Config
	HttpContext     echo.Context
	PasscodeService services.Passcode
	WebauthnService services.WebauthnService
	Persister       persistence.Persister
	SessionManager  session.Manager
	Tx              *pop.Connection
}

type Action struct{}

func (a *Action) GetDeps(c flowpilot.Context) *Dependencies {
	return c.Get("dependencies").(*Dependencies)
}
