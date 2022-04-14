package server

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/handler"
	"github.com/teamhanko/hanko/persistence"
)

func NewPublicRouter(cfg *config.Config, persister *persistence.Persister) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","time_unix":"${time_unix}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}},"referer":"${referer}"` + "\n",
	}))

	healthHandler := handler.NewHealthHandler()
	webauthnHandler, err := handler.NewWebauthnHandler(cfg.Webauthn, persister)
	if err != nil {
		panic(fmt.Errorf("failed to create public webauthn handler: %w", err))
	}

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	webauthn := e.Group("/webauthn")
	// TODO: webauthnRegistration must be protected with JWT/Session
	webauthnRegistration := webauthn.Group("/registration")
	webauthnRegistration.POST("/initialize", webauthnHandler.BeginRegistration)
	webauthnRegistration.POST("/finalize", webauthnHandler.FinishRegistration)

	webauthnLogin := webauthn.Group("/login")
	webauthnLogin.POST("/initialize", webauthnHandler.BeginAuthentication)
	webauthnLogin.POST("/finalize", webauthnHandler.FinishAuthentication)

	return e
}
