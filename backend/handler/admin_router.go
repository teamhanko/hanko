package handler

import (
	"fmt"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	hankoMiddleware "github.com/teamhanko/hanko/backend/middleware"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/template"
)

func NewAdminRouter(cfg *config.Config, persister persistence.Persister, prometheus echo.MiddlewareFunc) *echo.Echo {
	e := echo.New()
	e.Renderer = template.NewTemplateRenderer()
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
		e.Use(prometheus)
		e.GET("/metrics", echoprometheus.NewHandler())
	}

	statusHandler := NewStatusHandler(persister)

	e.GET("/", statusHandler.Status)

	healthHandler := NewHealthHandler()

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	userHandler := NewUserHandlerAdmin(persister)

	jwkManager, err := jwk.NewDefaultManager(cfg.Secrets.Keys, persister.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}

	webhookMiddleware := hankoMiddleware.WebhookMiddleware(cfg, jwkManager, persister.GetWebhookPersister(nil))

	user := g.Group("/users")
	user.GET("", userHandler.List)
	user.POST("", userHandler.Create, webhookMiddleware)
	user.GET("/:id", userHandler.Get)
	user.DELETE("/:id", userHandler.Delete, webhookMiddleware)

	auditLogHandler := NewAuditLogHandler(persister)

	auditLogs := g.Group("/audit_logs")
	auditLogs.GET("", auditLogHandler.List)

	webhookHandler := NewWebhookHandler(cfg.Webhooks, persister)
	webhooks := g.Group("/webhooks")
	webhooks.GET("", webhookHandler.List)
	webhooks.POST("", webhookHandler.Create)
	webhooks.GET("/:id", webhookHandler.Get)
	webhooks.DELETE("/:id", webhookHandler.Delete)
	webhooks.PUT("/:id", webhookHandler.Update)

	return e
}
