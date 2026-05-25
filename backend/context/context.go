package context

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v2/session"
)

// TenantIDFromContext returns the tenant ID stored in the Echo context.
//
// The context always contains a non-nil tenant ID. In single-tenant mode,
// it returns the reserved UUID (00000000-0000-0000-0000-000000000000). TODO: this is not good
// If the context contains a tenant value of an unexpected type, the function returns an error.
func TenantIDFromContext(c echo.Context) (uuid.UUID, error) {
	v := c.Get("tenant_id")
	if v == nil {
		return uuid.Nil, fmt.Errorf("tenant_id not found in context")
	}

	tenantID, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant_id has unexpected type %T", v)
	}

	return tenantID, nil
}

func GetTenant(c echo.Context) (*Tenant, error) {
	v := c.Get("tenant")
	if v == nil {
		return nil, fmt.Errorf("tenant not found in context")
	}

	tenant, ok := c.Get("tenant").(Tenant)
	if !ok {
		return nil, fmt.Errorf("tenant has unexpected type %T", c.Get("tenant"))
	}

	return &tenant, nil
}

func GetJwkManager(c echo.Context) (jwk.KeyProvider, error) {
	v := c.Get("jwk_manager")
	if v == nil {
		return nil, fmt.Errorf("tenant not found in context")
	}

	jwkManager, ok := c.Get("jwk_manager").(jwk.KeyProvider)
	if !ok {
		return nil, fmt.Errorf("JWK manager has unexpected type %T", c.Get("jwk_manager"))
	}

	return jwkManager, nil
}

func GetSessionManager(c echo.Context) (session.Manager, error) {
	v := c.Get("session_manager")
	if v == nil {
		return nil, fmt.Errorf("session manager not found in context")
	}

	sessionManager, ok := c.Get("session_manager").(session.Manager)
	if !ok {
		return nil, fmt.Errorf("JWK manager has unexpected type %T", c.Get("session_manager"))
	}

	return sessionManager, nil
}

func GetAppConfig(c echo.Context) (*config.ApplicationConfig, error) {
	v := c.Get("app_config")
	if v == nil {
		return nil, fmt.Errorf("application config not found in context")
	}

	appConfig, ok := c.Get("app_config").(config.ApplicationConfig)
	if !ok {
		return nil, fmt.Errorf("application config has unexpected type %T", c.Get("app_config"))
	}

	return &appConfig, nil
}
