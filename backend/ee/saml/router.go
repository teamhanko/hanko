package saml

import (
	"github.com/labstack/echo/v4"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/ee/saml/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
)

func CreateSamlRoutes(e *echo.Echo, cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) {
	handler := NewSamlHandler(cfg, persister, sessionManager, auditLogger)
	routingGroup := e.Group("saml")
	routingGroup.GET("/provider", handler.GetProvider)
	routingGroup.GET("/metadata", handler.Metadata)
	routingGroup.GET("/auth", handler.Auth)
	routingGroup.POST("/callback", handler.CallbackPost)
}

func CreateSamlAdminRoutes(e *echo.Echo, cfg *config.Config, persister persistence.Persister) {
	handler := admin.NewSamlAdminHandler(cfg, persister)

	routingGroup := e.Group("saml")
	routingGroup.GET("", handler.List)
	routingGroup.POST("", handler.Create)

	singleProviderGroup := routingGroup.Group("/:id")
	singleProviderGroup.GET("", handler.Get)
	singleProviderGroup.PUT("", handler.Update)
	singleProviderGroup.DELETE("", handler.Delete)
}
