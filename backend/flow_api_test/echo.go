package flow_api_test

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type FlowPilotHandler struct {
	Persister persistence.Persister
}

type FlowRequest struct {
	flowpilot.InputData
	FlowConfig
}

func (e *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	actionParam := c.QueryParam("flowpilot_action")

	var body FlowRequest
	_ = c.Bind(&body)

	flowConfig := FlowConfig{FlowOption: body.FlowOption}
	if !flowConfig.IsValid() {
		return fmt.Errorf("invalid flow option: %v", flowConfig)
	}

	myFlowConfig = flowConfig

	err := e.Persister.Transaction(func(tx *pop.Connection) error {
		db := models.NewFlowDB(tx)

		result, flowpilotErr := Flow.Execute(db,
			flowpilot.WithActionParam(actionParam),
			flowpilot.WithInputData(body.InputData))

		if flowpilotErr != nil {
			return flowpilotErr
		}

		return c.JSON(result.Status(), result.Response())
	})

	if err != nil {
		c.Logger().Errorf("tx error: %v", err)
		result := Flow.ResultFromError(err)

		return c.JSON(result.Status(), result.Response())
	}

	return nil
}
