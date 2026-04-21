package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/teamhanko/hanko/backend/v2/config"
)

// TenantAwareCORS returns a CORS middleware that dynamically validates origins based on tenant configuration.
// In multitenancy mode, it reads the tenant configuration from context (set by TenantMiddleware) to validate origins.
// In non-multitenant mode, it uses the default CORS configuration from the application config.
//
// This middleware should run AFTER TenantMiddleware, which loads tenant configuration into context.
// It delegates to Echo's CORSWithConfig which handles wildcard pattern matching and all CORS headers.
func TenantAwareCORS(multiTenancy bool, defaultCors config.Cors, exposeHeaders []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Determine which CORS config to use
			var corsConfig config.Cors

			if !multiTenancy {
				corsConfig = defaultCors
			} else {
				// Tenant config should already be loaded by TenantMiddleware
				tenantConfigRaw := c.Get("tenant_config")
				if tenantConfigRaw == nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "tenant config not loaded - TenantMiddleware must run before CORS")
				}
				tenantConfig, ok := tenantConfigRaw.(*config.TenantConfig)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, "invalid tenant_config type in context")
				}
				corsConfig = tenantConfig.Cors
			}

			// Delegate to Echo's CORS middleware
			// Echo handles:
			// - Wildcard pattern matching (* and ? wildcards converted to regex)
			// - All CORS headers (Access-Control-*)
			// - Preflight vs actual request handling
			// - UnsafeWildcardOriginWithAllowCredentials flag
			echoCorsConfig := echomiddleware.CORSConfig{
				AllowOrigins:                             corsConfig.AllowOrigins,
				AllowMethods:                             []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
				AllowCredentials:                         true,
				ExposeHeaders:                            exposeHeaders,
				MaxAge:                                   7200,
				UnsafeWildcardOriginWithAllowCredentials: corsConfig.UnsafeWildcardOriginAllowed,
			}

			return echomiddleware.CORSWithConfig(echoCorsConfig)(next)(c)
		}
	}
}
