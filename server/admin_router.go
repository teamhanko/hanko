package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/handler"
	"github.com/teamhanko/hanko/persistence"
)

func NewPrivateRouter(persister persistence.Persister) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","time_unix":"${time_unix}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}},"referer":"${referer}"` + "\n",
	}))

	e.Validator = dto.NewCustomValidator()

	healthHandler := handler.NewHealthHandler()

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	userHandler := handler.NewUserHandlerAdmin(persister)

	user := e.Group("/users")
	user.DELETE("/:id", userHandler.Delete)
	user.PATCH("/:id", userHandler.Patch)

	return e
}
