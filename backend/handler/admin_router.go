package handler

import (
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	hankoMiddleware "github.com/teamhanko/hanko/backend/v2/middleware"
	"github.com/teamhanko/hanko/backend/v2/persistence"
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

	jwkMiddleware := hankoMiddleware.JWKMiddleware(cfg.ApplicationConfig, persister)
	webhookMiddleware := hankoMiddleware.WebhookMiddleware(persister)
	sessionManagerMiddleware := hankoMiddleware.SessionManager()

	var tenantGroup *echo.Group
	var tenantMiddleware echo.MiddlewareFunc
	if cfg.MultiTenancy {
		tenantMiddleware = hankoMiddleware.TenantMiddlewareMultitenancy(persister)
		tenantGroup = g.Group("/:tenant_id", tenantMiddleware)
	} else {
		tenantMiddleware = hankoMiddleware.TenantMiddlewareSingleTenant(cfg.TenantConfig)
		tenantGroup = g.Group("", tenantMiddleware)
	}

	auditLogger := auditlog.NewLogger(persister, cfg.AuditLog)

	userHandler := NewUserHandlerAdmin(persister)
	metadataHandler := NewMetadataAdminHandler(persister)
	emailHandler := NewEmailAdminHandler(persister)
	webauthnCredentialHandler := NewWebauthnCredentialAdminHandler(persister)
	passwordCredentialHandler := NewPasswordAdminHandler(persister)
	sessionsHandler := NewSessionAdminHandler(persister, auditLogger)
	webhookHandler := NewWebhookHandler(cfg.Webhooks, persister)

	user := tenantGroup.Group("/users")
	user.GET("", userHandler.List)
	user.POST("", userHandler.Create, jwkMiddleware, webhookMiddleware)
	user.GET("/:id", userHandler.Get)
	user.DELETE("/:id", userHandler.Delete, jwkMiddleware, webhookMiddleware)
	user.PATCH("/:id", userHandler.Patch)

	user.PATCH("/:id/metadata", metadataHandler.PatchMetadata)
	user.GET("/:id/metadata", metadataHandler.GetMetadata)

	email := user.Group("/:user_id/emails", jwkMiddleware, webhookMiddleware)
	email.GET("", emailHandler.List)
	email.POST("", emailHandler.Create)
	email.GET("/:email_id", emailHandler.Get)
	email.DELETE("/:email_id", emailHandler.Delete)
	email.POST("/:email_id/set_primary", emailHandler.SetPrimaryEmail)

	webauthnCredentials := user.Group("/:user_id/webauthn_credentials")
	webauthnCredentials.GET("", webauthnCredentialHandler.List)
	webauthnCredentials.GET("/:credential_id", webauthnCredentialHandler.Get)
	webauthnCredentials.DELETE("/:credential_id", webauthnCredentialHandler.Delete)

	passwordCredentials := user.Group("/:user_id/password")
	passwordCredentials.GET("", passwordCredentialHandler.Get)
	passwordCredentials.POST("", passwordCredentialHandler.Create)
	passwordCredentials.PUT("", passwordCredentialHandler.Update)
	passwordCredentials.DELETE("", passwordCredentialHandler.Delete)

	userSessions := user.Group("/:user_id/sessions", jwkMiddleware, sessionManagerMiddleware)
	userSessions.GET("", sessionsHandler.List)
	userSessions.DELETE("/:session_id", sessionsHandler.Delete)

	otpHandler := NewOTPAdminHandler(persister)
	otp := user.Group("/:user_id/otp")
	otp.GET("", otpHandler.Get)
	otp.DELETE("", otpHandler.Delete)

	auditLogHandler := NewAuditLogHandler(persister)

	auditLogs := tenantGroup.Group("/audit_logs")
	auditLogs.GET("", auditLogHandler.List)

	webhooks := tenantGroup.Group("/webhooks")
	webhooks.GET("", webhookHandler.List)
	webhooks.POST("", webhookHandler.Create)
	webhooks.GET("/:id", webhookHandler.Get)
	webhooks.DELETE("/:id", webhookHandler.Delete)
	webhooks.PUT("/:id", webhookHandler.Update)

	sessions := tenantGroup.Group("/sessions", jwkMiddleware, sessionManagerMiddleware)
	sessions.POST("", sessionsHandler.Generate)

	return e
}
