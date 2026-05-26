package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

func NewManagementRouter(cfg *config.Config, persister persistence.Persister) *echo.Echo {

	e := echo.New()
	e.HideBanner = true

	e.HTTPErrorHandler = dto.NewHTTPErrorHandler(dto.HTTPErrorHandlerConfig{Debug: false, Logger: e.Logger})
	e.Use(middleware.RequestID())

	e.Validator = dto.NewCustomValidator()

	statusHandler := NewStatusHandler(persister)
	e.GET("/", statusHandler.Status)

	healthHandler := NewHealthHandler()

	health := e.Group("/health")
	health.GET("/alive", healthHandler.Alive)
	health.GET("/ready", healthHandler.Ready)

	tenantHandler := NewTenantHandler(cfg, persister)

	tenants := e.Group("/tenants")
	tenants.POST("", tenantHandler.Create)
	tenants.GET("", tenantHandler.List)
	tenants.GET("/:id", tenantHandler.Get)
	tenants.PUT("/:id", tenantHandler.Update)
	tenants.DELETE("/:id", tenantHandler.Delete)

	samlProviderHandler := NewSamlProviderHandler(cfg, persister)

	saml := tenants.Group("/:tenantId/saml")
	samlProviders := saml.Group("/providers")
	samlProviders.POST("", samlProviderHandler.Create)
	samlProviders.GET("", samlProviderHandler.List)
	samlProviders.GET("/:providerId", samlProviderHandler.Get)
	samlProviders.PUT("/:providerId", samlProviderHandler.Update)
	samlProviders.DELETE("/:providerId", samlProviderHandler.Delete)

	return e
}
