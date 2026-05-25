package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/teamhanko/hanko/backend/v2/context"
)

func TenantAwareCORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant, err := context.GetTenant(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get tenant from context - TenantMiddleware must run before CORS")
			}

			corsConfig := tenant.Config.Cors

			exposeHeaders := []string{
				httplimit.HeaderRetryAfter,
				httplimit.HeaderRateLimitLimit,
				httplimit.HeaderRateLimitRemaining,
				httplimit.HeaderRateLimitReset,
				"X-Session-Lifetime",
				"X-Session-Retention",
			}

			if tenant.Config.Session.EnableAuthTokenHeader {
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
