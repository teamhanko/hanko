package server

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/handler"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	hankoMiddleware "github.com/teamhanko/hanko/backend/server/middleware"
	"github.com/teamhanko/hanko/backend/session"
)

func NewPublicRouter(cfg *config.Config, persister persistence.Persister) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.HTTPErrorHandler = dto.NewHTTPErrorHandler(dto.HTTPErrorHandlerConfig{Debug: false, Logger: e.Logger})
	e.Use(middleware.RequestID())
	e.Use(hankoMiddleware.GetLoggerMiddleware())

	if cfg.Server.Public.Cors.Enabled {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.Server.Public.Cors.AllowOrigins,
			AllowMethods:     cfg.Server.Public.Cors.AllowMethods,
			AllowHeaders:     cfg.Server.Public.Cors.AllowHeaders,
			ExposeHeaders:    cfg.Server.Public.Cors.ExposeHeaders,
			AllowCredentials: cfg.Server.Public.Cors.AllowCredentials,
			MaxAge:           cfg.Server.Public.Cors.MaxAge,
		}))
	}

	e.Validator = dto.NewCustomValidator()

	jwkManager, err := jwk.NewDefaultManager(cfg.Secrets.Keys, persister.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, cfg.Session)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}

	mailer, err := mail.NewMailer(cfg.Passcode.Smtp)
	if err != nil {
		panic(fmt.Errorf("failed to create mailer: %w", err))
	}

	auditLogger := auditlog.NewLogger(persister, cfg.AuditLog)

	if cfg.Password.Enabled {
		passwordHandler := handler.NewPasswordHandler(persister, sessionManager, cfg, auditLogger)

		password := e.Group("/password")
		password.PUT("", passwordHandler.Set, hankoMiddleware.Session(sessionManager))
		password.POST("/login", passwordHandler.Login)
	}

	userHandler := handler.NewUserHandler(cfg, persister, auditLogger)

	e.GET("/me", userHandler.Me, hankoMiddleware.Session(sessionManager))

	user := e.Group("/users")
	user.POST("", userHandler.Create)
	user.GET("/:id", userHandler.Get, hankoMiddleware.Session(sessionManager))

	e.POST("/user", userHandler.GetUserIdByEmail)

	healthHandler := handler.NewHealthHandler()
	webauthnHandler, err := handler.NewWebauthnHandler(cfg, persister, sessionManager, auditLogger)
	if err != nil {
		panic(fmt.Errorf("failed to create public webauthn handler: %w", err))
	}
	passcodeHandler, err := handler.NewPasscodeHandler(cfg, persister, sessionManager, mailer, auditLogger)
	if err != nil {
		panic(fmt.Errorf("failed to create public passcode handler: %w", err))
	}

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	wellKnownHandler, err := handler.NewWellKnownHandler(*cfg, jwkManager)
	if err != nil {
		panic(fmt.Errorf("failed to create well-known handler: %w", err))
	}
	wellKnown := e.Group("/.well-known")
	wellKnown.GET("/jwks.json", wellKnownHandler.GetPublicKeys)
	wellKnown.GET("/config", wellKnownHandler.GetConfig)

	webauthn := e.Group("/webauthn")
	webauthnRegistration := webauthn.Group("/registration", hankoMiddleware.Session(sessionManager))
	webauthnRegistration.POST("/initialize", webauthnHandler.BeginRegistration)
	webauthnRegistration.POST("/finalize", webauthnHandler.FinishRegistration)

	webauthnLogin := webauthn.Group("/login")
	webauthnLogin.POST("/initialize", webauthnHandler.BeginAuthentication)
	webauthnLogin.POST("/finalize", webauthnHandler.FinishAuthentication)

	passcode := e.Group("/passcode")
	passcodeLogin := passcode.Group("/login")
	passcodeLogin.POST("/initialize", passcodeHandler.Init)
	passcodeLogin.POST("/finalize", passcodeHandler.Finish)

	return e
}
