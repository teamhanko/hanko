package handler

import (
	"fmt"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v2/dto"
	hankoMiddleware "github.com/teamhanko/hanko/backend/v2/middleware"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/session"
	"github.com/teamhanko/hanko/backend/v2/template"
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

	jwkManager, err := jwk.NewManager(cfg.Secrets, persister)
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, *cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}

	webhookMiddleware := hankoMiddleware.WebhookMiddleware(cfg, jwkManager, persister)
	auditLogger := auditlog.NewLogger(persister, cfg.AuditLog)

	userHandler := NewUserHandlerAdmin(persister)
	emailHandler := NewEmailAdminHandler(cfg, persister)
	sessionsHandler := NewSessionAdminHandler(cfg, persister, sessionManager, auditLogger)

	user := g.Group("/users")
	user.GET("", userHandler.List)
	user.POST("", userHandler.Create, webhookMiddleware)
	user.GET("/:id", userHandler.Get)
	user.DELETE("/:id", userHandler.Delete, webhookMiddleware)

	metadataHandler := NewMetadataAdminHandler(persister)
	user.PATCH("/:id/metadata", metadataHandler.PatchMetadata)
	user.GET("/:id/metadata", metadataHandler.GetMetadata)

	email := user.Group("/:user_id/emails", webhookMiddleware)
	email.GET("", emailHandler.List)
	email.POST("", emailHandler.Create)
	email.GET("/:email_id", emailHandler.Get)
	email.DELETE("/:email_id", emailHandler.Delete)
	email.POST("/:email_id/set_primary", emailHandler.SetPrimaryEmail)

	webauthnCredentialHandler := NewWebauthnCredentialAdminHandler(persister)
	webauthnCredentials := user.Group("/:user_id/webauthn_credentials")
	webauthnCredentials.GET("", webauthnCredentialHandler.List)
	webauthnCredentials.GET("/:credential_id", webauthnCredentialHandler.Get)
	webauthnCredentials.DELETE("/:credential_id", webauthnCredentialHandler.Delete)

	passwordCredentialHandler := NewPasswordAdminHandler(persister)
	passwordCredentials := user.Group("/:user_id/password")
	passwordCredentials.GET("", passwordCredentialHandler.Get)
	passwordCredentials.POST("", passwordCredentialHandler.Create)
	passwordCredentials.PUT("", passwordCredentialHandler.Update)
	passwordCredentials.DELETE("", passwordCredentialHandler.Delete)

	userSessions := user.Group("/:user_id/sessions")
	userSessions.GET("", sessionsHandler.List)
	userSessions.DELETE("/:session_id", sessionsHandler.Delete)

	otpHandler := NewOTPAdminHandler(persister)
	otp := user.Group("/:user_id/otp")
	otp.GET("", otpHandler.Get)
	otp.DELETE("", otpHandler.Delete)

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

	sessions := g.Group("/sessions")
	sessions.POST("", sessionsHandler.Generate)

	// Tenant management routes (only available when multi-tenant mode can be enabled)
	tenantHandler := NewTenantHandlerAdmin(persister)
	tenants := g.Group("/tenants")
	tenants.GET("", tenantHandler.List)
	tenants.POST("", tenantHandler.Create, webhookMiddleware)
	tenants.GET("/:id", tenantHandler.Get)
	tenants.PUT("/:id", tenantHandler.Update, webhookMiddleware)
	tenants.DELETE("/:id", tenantHandler.Delete, webhookMiddleware)

	return e
}
