package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/httplimit"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/ee/saml"
	"github.com/teamhanko/hanko/backend/v2/flow_api"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow_locker"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/mapper"
	hankoMiddleware "github.com/teamhanko/hanko/backend/v2/middleware"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/rate_limiter"
	"github.com/teamhanko/hanko/backend/v2/session"
	"github.com/teamhanko/hanko/backend/v2/template"
)

func NewPublicRouter(cfg *config.Config, persister persistence.Persister, prometheus echo.MiddlewareFunc, authenticatorMetadata mapper.AuthenticatorMetadata) *echo.Echo {
	e := echo.New()

	e.Renderer = template.NewTemplateRenderer()

	e.Static("/flowpilot", "flow_api/static") // TODO: remove!

	auditLogger := auditlog.NewLogger(persister, cfg.AuditLog)

	emailService, _ := services.NewEmailService()
	passcodeService := services.NewPasscodeService(*emailService, persister)
	passwordService := services.NewPasswordService(persister)
	webauthnService := services.NewWebauthnService(*cfg, persister)
	securityNotificationService := services.NewSecurityNotificationService(*emailService, persister, auditLogger)

	jwkManager, err := jwk.NewManager(cfg.Secrets, persister)
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, *cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}

	var otpRateLimiter limiter.Store
	var passcodeRateLimiter limiter.Store
	var passwordRateLimiter limiter.Store
	var tokenExchangeRateLimiter limiter.Store
	if cfg.RateLimiter.Enabled {
		otpRateLimiter = rate_limiter.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.OTPLimits)
		passcodeRateLimiter = rate_limiter.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.PasscodeLimits)
		passwordRateLimiter = rate_limiter.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.PasswordLimits)
		tokenExchangeRateLimiter = rate_limiter.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.TokenLimits)
	}

	samlService := saml.NewSamlService(cfg, persister)

	flowAPIHandler := flow_api.FlowPilotHandler{
		Persister:                   persister,
		Cfg:                         *cfg,
		PasscodeService:             passcodeService,
		SecurityNotificationService: securityNotificationService,
		PasswordService:             passwordService,
		WebauthnService:             webauthnService,
		SessionManager:              sessionManager,
		OTPRateLimiter:              otpRateLimiter,
		PasscodeRateLimiter:         passcodeRateLimiter,
		PasswordRateLimiter:         passwordRateLimiter,
		TokenExchangeRateLimiter:    tokenExchangeRateLimiter,
		AuthenticatorMetadata:       authenticatorMetadata,
		AuditLogger:                 auditLogger,
		SamlService:                 samlService,
	}

	flowLocker, err := flow_locker.NewFlowLocker(cfg.FlowLocker)
	if err != nil {
		panic(fmt.Errorf("failed to initialize flow locker: %w", err))
	}
	flowAPIHandler.FlowLocker = flowLocker

	if cfg.Saml.Enabled {
		saml.CreateSamlRoutes(e, sessionManager, auditLogger, samlService)
	}

	sessionMiddleware := hankoMiddleware.Session(cfg, persister, sessionManager)

	webhookMiddleware := hankoMiddleware.WebhookMiddleware(cfg, jwkManager, persister)
	tenantMiddleware := hankoMiddleware.TenantMiddleware(cfg.MultiTenancy, &cfg.TenantConfig, persister)

	var g *echo.Group
	if cfg.MultiTenancy {
		g = e.Group("/:tenant_id")
	} else {
		g = e.Group("")
	}

	g.POST("/registration", flowAPIHandler.RegistrationFlowHandler, tenantMiddleware, webhookMiddleware)
	g.POST("/login", flowAPIHandler.LoginFlowHandler, tenantMiddleware, webhookMiddleware)
	g.POST("/profile", flowAPIHandler.ProfileFlowHandler, tenantMiddleware, webhookMiddleware)

	if cfg.Saml.Enabled {
		e.POST("/token_exchange", flowAPIHandler.TokenExchangeFlowHandler, webhookMiddleware)
	}

	e.HideBanner = true

	e.HTTPErrorHandler = dto.NewHTTPErrorHandler(dto.HTTPErrorHandlerConfig{Debug: false, Logger: e.Logger})
	e.Use(middleware.RequestID())
	if cfg.Log.LogHealthAndMetrics {
		e.Use(hankoMiddleware.GetLoggerMiddleware())
	} else {
		g.Use(hankoMiddleware.GetLoggerMiddleware())
	}

	exposeHeader := []string{
		httplimit.HeaderRetryAfter,
		httplimit.HeaderRateLimitLimit,
		httplimit.HeaderRateLimitRemaining,
		httplimit.HeaderRateLimitReset,
		"X-Session-Lifetime",
		"X-Session-Retention",
	}

	if cfg.Session.EnableAuthTokenHeader {
		exposeHeader = append(exposeHeader, "X-Auth-Token")
	}

	// TODO: CORS origins are tenant specific
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		UnsafeWildcardOriginWithAllowCredentials: cfg.Server.Public.Cors.UnsafeWildcardOriginAllowed,
		AllowOrigins:                             cfg.Server.Public.Cors.AllowOrigins,
		ExposeHeaders:                            exposeHeader,
		AllowCredentials:                         true,
		// Based on: Chromium (starting in v76) caps at 2 hours (7200 seconds).
		MaxAge: 7200,
	}))

	if prometheus != nil {
		e.Use(prometheus)
	}

	e.Validator = dto.NewCustomValidator()

	userHandler := NewUserHandler(cfg, persister, sessionManager, auditLogger)
	statusHandler := NewStatusHandler(persister)

	e.GET("/", statusHandler.Status)
	g.GET("/me", userHandler.Me, tenantMiddleware, sessionMiddleware)

	g.POST("/logout", userHandler.Logout, tenantMiddleware, sessionMiddleware)

	healthHandler := NewHealthHandler()

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	wellKnownHandler, err := NewWellKnownHandler(*cfg, jwkManager)
	if err != nil {
		panic(fmt.Errorf("failed to create well-known handler: %w", err))
	}
	wellKnown := g.Group("/.well-known")
	wellKnown.GET("/jwks.json", wellKnownHandler.GetPublicKeys)

	thirdPartyHandler := NewThirdPartyHandler(cfg, persister, sessionManager, auditLogger)
	thirdparty := g.Group("thirdparty", tenantMiddleware)
	thirdparty.GET("/callback", thirdPartyHandler.Callback, webhookMiddleware)
	thirdparty.POST("/callback", thirdPartyHandler.CallbackPost, webhookMiddleware)

	sessionHandler := NewSessionHandler(persister, sessionManager, *cfg)
	sessions := g.Group("sessions", tenantMiddleware)
	sessions.GET("/validate", sessionHandler.ValidateSession)
	sessions.POST("/validate", sessionHandler.ValidateSessionFromBody)

	return e
}
