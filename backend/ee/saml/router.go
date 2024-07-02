package saml

import (
	"github.com/labstack/echo/v4"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/session"
)

func CreateSamlRoutes(e *echo.Echo, sessionManager session.Manager, auditLogger auditlog.Logger, samlService Service) {
	handler := NewSamlHandler(sessionManager, auditLogger, samlService)
	routingGroup := e.Group("saml")
	routingGroup.GET("/provider", handler.GetProvider)
	routingGroup.GET("/metadata", handler.Metadata)
	routingGroup.GET("/auth", handler.Auth)
	routingGroup.POST("/callback", handler.CallbackPost)
}
