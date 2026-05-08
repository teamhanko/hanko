package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/teamhanko/hanko/backend/v3/config"
)

func TenantAwareCORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Tenant config should already be loaded by TenantMiddleware
			tenantConfigRaw := c.Get("tenant_config")
			if tenantConfigRaw == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "tenant config not loaded - TenantMiddleware must run before CORS")
			}

			tenantConfig, ok := tenantConfigRaw.(*config.TenantConfig)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "invalid tenant_config type in context")
			}

			corsConfig := tenantConfig.Cors

			exposeHeaders := []string{
				httplimit.HeaderRetryAfter,
				httplimit.HeaderRateLimitLimit,
				httplimit.HeaderRateLimitRemaining,
				httplimit.HeaderRateLimitReset,
				"X-Session-Lifetime",
				"X-Session-Retention",
			}

			if tenantConfig.Session.EnableAuthTokenHeader {
				exposeHeaders = append(exposeHeaders, "X-Auth-Token")
			}

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
