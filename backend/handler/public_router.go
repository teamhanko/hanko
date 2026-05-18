package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/flow_api"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow_locker"
	"github.com/teamhanko/hanko/backend/v2/mapper"
	hankoMiddleware "github.com/teamhanko/hanko/backend/v2/middleware"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/template"
)

func NewPublicRouter(cfg *config.Config, persister persistence.Persister, prometheus echo.MiddlewareFunc, authenticatorMetadata mapper.AuthenticatorMetadata) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Renderer = template.NewTemplateRenderer()
	e.Validator = dto.NewCustomValidator()
	e.HTTPErrorHandler = dto.NewHTTPErrorHandler(dto.HTTPErrorHandlerConfig{Debug: cfg.Debug, Logger: e.Logger})

	g := e.Group("")
	if cfg.Log.LogHealthAndMetrics {
		e.Use(hankoMiddleware.GetLoggerMiddleware())
	} else {
		g.Use(hankoMiddleware.GetLoggerMiddleware())
	}

	if prometheus != nil {
		e.Use(prometheus)
	}

	if cfg.Log.LogHealthAndMetrics {
		e.Use(hankoMiddleware.GetLoggerMiddleware())
	}

	e.Use(middleware.RequestID())
	e.Static("/flowpilot", "flow_api/static") // TODO: remove!

	auditLogger := auditlog.NewLogger(persister, cfg.AuditLog)

	flowAPIHandler := flow_api.NewFlowAPIHandler(*cfg, persister, auditLogger, authenticatorMetadata)

	flowLocker, err := flow_locker.NewFlowLocker(cfg.FlowLocker)
	if err != nil {
		panic(fmt.Errorf("failed to initialize flow locker: %w", err))
	}
	flowAPIHandler.FlowLocker = flowLocker

	sessionMiddleware := hankoMiddleware.Session(persister)
	webhookMiddleware := hankoMiddleware.WebhookMiddleware(persister)
	corsMiddleware := hankoMiddleware.TenantAwareCORS()
	jwkMiddleware := hankoMiddleware.JWKMiddleware(cfg.ApplicationConfig, persister)
	sessionManagerMiddleware := hankoMiddleware.SessionManager()

	var tenantGroupRoot *echo.Group
	if cfg.MultiTenancy.Enabled {
		tenantMiddleware := hankoMiddleware.TenantMiddlewareMultitenancy(persister)
		tenantGroupRoot = e.Group("/:tenant_id", tenantMiddleware)
	} else {
		tenantMiddleware := hankoMiddleware.TenantMiddlewareSingleTenant(cfg.TenantConfig)
		tenantGroupRoot = e.Group("", tenantMiddleware)
	}

	tenantGroup := tenantGroupRoot.Group("", corsMiddleware, jwkMiddleware, sessionManagerMiddleware)

	userHandler := NewUserHandler(persister, auditLogger)
	statusHandler := NewStatusHandler(persister)
	healthHandler := NewHealthHandler()

	tenantGroup.POST("/registration", flowAPIHandler.RegistrationFlowHandler, webhookMiddleware)
	tenantGroup.POST("/login", flowAPIHandler.LoginFlowHandler, webhookMiddleware)
	tenantGroup.POST("/profile", flowAPIHandler.ProfileFlowHandler, webhookMiddleware)

	//if cfg.Saml.Enabled {
	//	samlHandler := saml.NewSamlHandler(auditLogger, samlService)
	//	samlGroup := tenantGroup.Group("/saml")
	//	samlGroup.GET("/metadata", samlHandler.Metadata)
	//	samlGroup.GET("/auth", samlHandler.Auth)
	//	samlGroup.POST("/callback", samlHandler.CallbackPost)
	//	tenantGroup.POST("/token_exchange", flowAPIHandler.TokenExchangeFlowHandler, webhookMiddleware)
	//}

	tenantGroup.GET("/", statusHandler.Status)
	tenantGroup.GET("/me", userHandler.Me, sessionMiddleware)
	tenantGroup.POST("/logout", userHandler.Logout, sessionMiddleware)

	health := tenantGroup.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	wellKnownHandler, err := NewWellKnownHandler()
	if err != nil {
		panic(fmt.Errorf("failed to create well-known handler: %w", err))
	}
	wellKnown := tenantGroup.Group("/.well-known")
	wellKnown.GET("/jwks.json", wellKnownHandler.GetPublicKeys)

	thirdPartyHandler := NewThirdPartyHandler(persister, auditLogger)
	thirdparty := tenantGroup.Group("/thirdparty")
	thirdparty.GET("/callback", thirdPartyHandler.Callback, webhookMiddleware)
	thirdparty.POST("/callback", thirdPartyHandler.CallbackPost, webhookMiddleware)

	sessionHandler := NewSessionHandler(persister)
	sessions := tenantGroup.Group("/sessions")
	sessions.GET("/validate", sessionHandler.ValidateSession)
	sessions.POST("/validate", sessionHandler.ValidateSessionFromBody)

	return e
}
