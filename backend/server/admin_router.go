package server

import (
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/handler"
	"github.com/teamhanko/hanko/backend/persistence"
	hankoMiddleware "github.com/teamhanko/hanko/backend/server/middleware"
)

func NewAdminRouter(cfg *config.Config, persister persistence.Persister, prometheus *prometheus.Prometheus) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	g := e.Group("")

	e.HTTPErrorHandler = dto.NewHTTPErrorHandler(dto.HTTPErrorHandlerConfig{Debug: false, Logger: e.Logger})
	e.Use(middleware.RequestID())
	if cfg.Log.LogHealthAndMetrics {
		e.Use(hankoMiddleware.GetLoggerMiddleware())
	} else {
		g.Use(hankoMiddleware.GetLoggerMiddleware())
	}

	e.Validator = dto.NewCustomValidator()

	if prometheus != nil {
		prometheus.SetMetricsPath(e)
	}

	healthHandler := handler.NewHealthHandler()

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	userHandler := handler.NewUserHandlerAdmin(persister)

	user := g.Group("/users")
	user.GET("", userHandler.List)
	user.GET("/:id", userHandler.Get)
	user.DELETE("/:id", userHandler.Delete)

	auditLogHandler := handler.NewAuditLogHandler(persister)

	auditLogs := g.Group("/audit_logs")
	auditLogs.GET("", auditLogHandler.List)

	return e
}
