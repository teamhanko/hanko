package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/handler"
	"github.com/teamhanko/hanko/backend/persistence"
	hankoMiddleware "github.com/teamhanko/hanko/backend/server/middleware"
)

func NewAdminRouter(persister persistence.Persister) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestID())
	e.Use(hankoMiddleware.GetLoggerMiddleware())

	e.Validator = dto.NewCustomValidator()

	healthHandler := handler.NewHealthHandler()

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	userHandler := handler.NewUserHandlerAdmin(persister)

	user := e.Group("/users")
	user.DELETE("/:id", userHandler.Delete)
	user.PATCH("/:id", userHandler.Patch)
	user.GET("", userHandler.List)

	auditLogHandler := handler.NewAuditLogHandler(persister)

	auditLogs := e.Group("/audit_logs")
	auditLogs.GET("", auditLogHandler.List)

	return e
}
