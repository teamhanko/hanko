package handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/ee/saml"
	"github.com/teamhanko/hanko/backend/flow_api"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/mapper"
	hankoMiddleware "github.com/teamhanko/hanko/backend/middleware"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/rate_limiter"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/template"
)

func NewPublicRouter(cfg *config.Config, persister persistence.Persister, prometheus echo.MiddlewareFunc, authenticatorMetadata mapper.AuthenticatorMetadata) *echo.Echo {
	e := echo.New()

	e.Renderer = template.NewTemplateRenderer()

	e.Static("/flowpilot", "flow_api/static") // TODO: remove!

	emailService, err := services.NewEmailService(*cfg)
	passcodeService := services.NewPasscodeService(*cfg, *emailService, persister)
	passwordService := services.NewPasswordService(persister)
	webauthnService := services.NewWebauthnService(*cfg, persister)

	jwkManager, err := jwk.NewDefaultManager(cfg.Secrets.Keys, persister.GetJwkPersister())
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

	auditLogger := auditlog.NewLogger(persister, cfg.AuditLog)

	samlService := saml.NewSamlService(cfg, persister)

	flowAPIHandler := flow_api.FlowPilotHandler{
		Persister:                persister,
		Cfg:                      *cfg,
		PasscodeService:          passcodeService,
		PasswordService:          passwordService,
		WebauthnService:          webauthnService,
		SessionManager:           sessionManager,
		OTPRateLimiter:           otpRateLimiter,
		PasscodeRateLimiter:      passcodeRateLimiter,
		PasswordRateLimiter:      passwordRateLimiter,
		TokenExchangeRateLimiter: tokenExchangeRateLimiter,
		AuthenticatorMetadata:    authenticatorMetadata,
		AuditLogger:              auditLogger,
		SamlService:              samlService,
	}

	if cfg.Saml.Enabled {
		saml.CreateSamlRoutes(e, sessionManager, auditLogger, samlService)
	}

	sessionMiddleware := hankoMiddleware.Session(cfg, persister, sessionManager)

	webhookMiddleware := hankoMiddleware.WebhookMiddleware(cfg, jwkManager, persister)

	e.POST("/registration", flowAPIHandler.RegistrationFlowHandler, webhookMiddleware)
	e.POST("/login", flowAPIHandler.LoginFlowHandler, webhookMiddleware)
	e.POST("/profile", flowAPIHandler.ProfileFlowHandler, webhookMiddleware)

	if cfg.Saml.Enabled {
		e.POST("/token_exchange", flowAPIHandler.TokenExchangeFlowHandler, webhookMiddleware)
	}

	e.HideBanner = true
	g := e.Group("")

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

	mailer, err := mail.NewMailer(cfg.EmailDelivery.SMTP)
	if err != nil {
		panic(fmt.Errorf("failed to create mailer: %w", err))
	}

	if !cfg.MFA.Enabled && cfg.Password.Enabled {
		passwordHandler := NewPasswordHandler(persister, sessionManager, cfg, auditLogger)

		password := g.Group("/password")
		password.PUT("", passwordHandler.Set, sessionMiddleware)
		password.POST("/login", passwordHandler.Login)
	}

	userHandler := NewUserHandler(cfg, persister, sessionManager, auditLogger)
	statusHandler := NewStatusHandler(persister)

	e.GET("/", statusHandler.Status)
	g.GET("/me", userHandler.Me, sessionMiddleware)

	user := g.Group("/users", webhookMiddleware)
	user.POST("", userHandler.Create)
	user.GET("/:id", userHandler.Get, sessionMiddleware)

	g.POST("/user", userHandler.GetUserIdByEmail)
	g.POST("/logout", userHandler.Logout, sessionMiddleware)

	if cfg.Account.AllowDeletion {
		g.DELETE("/user", userHandler.Delete, sessionMiddleware, webhookMiddleware)
	}

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
	wellKnown.GET("/config", wellKnownHandler.GetConfig)

	emailHandler := NewEmailHandler(cfg, persister, auditLogger)

	if cfg.Passkey.Enabled {
		webauthnHandler, err := NewWebauthnHandler(cfg, persister, sessionManager, auditLogger, authenticatorMetadata)
		if err != nil {
			panic(fmt.Errorf("failed to create public webauthn handler: %w", err))
		}
		webauthn := g.Group("/webauthn")
		webauthnRegistration := webauthn.Group("/registration", sessionMiddleware)
		webauthnRegistration.POST("/initialize", webauthnHandler.BeginRegistration)
		webauthnRegistration.POST("/finalize", webauthnHandler.FinishRegistration)

		webauthnLogin := webauthn.Group("/login")
		webauthnLogin.POST("/initialize", webauthnHandler.BeginAuthentication)
		webauthnLogin.POST("/finalize", webauthnHandler.FinishAuthentication)

		webauthnCredentials := webauthn.Group("/credentials", sessionMiddleware)
		webauthnCredentials.GET("", webauthnHandler.ListCredentials)
		webauthnCredentials.PATCH("/:id", webauthnHandler.UpdateCredential)
		webauthnCredentials.DELETE("/:id", webauthnHandler.DeleteCredential)
	}

	if !cfg.MFA.Enabled && cfg.Email.Enabled && cfg.Email.UseForAuthentication {
		passcodeHandler, err := NewPasscodeHandler(cfg, persister, sessionManager, mailer, auditLogger)
		if err != nil {
			panic(fmt.Errorf("failed to create public passcode handler: %w", err))
		}
		passcode := g.Group("/passcode")
		passcodeLogin := passcode.Group("/login", webhookMiddleware)
		passcodeLogin.POST("/initialize", passcodeHandler.Init)
		passcodeLogin.POST("/finalize", passcodeHandler.Finish)
	}

	email := g.Group("/emails", sessionMiddleware, webhookMiddleware)
	email.GET("", emailHandler.List)
	email.POST("", emailHandler.Create)
	email.DELETE("/:id", emailHandler.Delete)
	email.POST("/:id/set_primary", emailHandler.SetPrimaryEmail)

	thirdPartyHandler := NewThirdPartyHandler(cfg, persister, sessionManager, auditLogger)
	thirdparty := g.Group("thirdparty")
	thirdparty.GET("/auth", thirdPartyHandler.Auth)
	thirdparty.GET("/callback", thirdPartyHandler.Callback, webhookMiddleware)
	thirdparty.POST("/callback", thirdPartyHandler.CallbackPost, webhookMiddleware)

	tokenHandler := NewTokenHandler(cfg, persister, sessionManager, auditLogger)
	g.POST("/token", tokenHandler.Validate)

	sessionHandler := NewSessionHandler(persister, sessionManager, *cfg)
	sessions := g.Group("sessions")
	sessions.GET("/validate", sessionHandler.ValidateSession)
	sessions.POST("/validate", sessionHandler.ValidateSessionFromBody)

	return e
}
