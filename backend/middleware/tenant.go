package middleware

import (
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

// TenantMiddleware is a middleware function that retrieves tenant information using the tenant ID from the request path.
// It loads the tenant configuration and sets it on the context for downstream middlewares and handlers to use.
// This middleware should run early in the middleware chain, before other middlewares that need tenant context (like CORS).
// TODO: maybe return flowErrors instead
func TenantMiddleware(multiTenancy bool, tenantConfig *config.TenantConfig, persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Skip if tenant config already loaded (avoid duplicate work)
			if ctx.Get("tenant_config") != nil {
				return next(ctx)
			}

			if multiTenancy {
				// Try to get tenant_id from route param first, fallback to path extraction
				var tenantIdStr string
				tenantIdStr = ctx.Param("tenant_id")
				if tenantIdStr == "" {
					// Route param not available yet (e.g., early in middleware chain)
					// Extract from path directly
					tenantIdStr = extractTenantIDFromPath(ctx.Request().URL.Path)
				}

				if tenantIdStr == "" {
					return echo.NewHTTPError(http.StatusBadRequest, "tenant ID required")
				}

				tenantId, err := uuid.FromString(tenantIdStr)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID format").SetInternal(err)
				}

				tenant, err := persister.GetTenantPersister().Get(tenantId)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tenant").SetInternal(err)
				}
				if tenant == nil {
					return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
				}

				defaultTenantConfig := config.DefaultTenantConfig()
				tenantConfig = &defaultTenantConfig
				k := koanf.New(".")

				if err := k.Load(rawbytes.Provider(tenant.Config), json.Parser()); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse tenant config").SetInternal(err)
				}

				if err := k.Unmarshal("", tenantConfig); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to unmarshal tenant config").SetInternal(err)
				}

				if err := tenantConfig.Webauthn.PostProcess(); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to post process tenant settings").SetInternal(err)
				}

				ctx.Set("tenant_config", tenantConfig)
				ctx.Set("tenant_id", &tenantId)
			} else {
				ctx.Set("tenant_id", nil)
				ctx.Set("tenant_config", tenantConfig)
			}
			return next(ctx)
		}
	}
}

// extractTenantIDFromPath extracts the tenant ID from the URL path
// Expected format: /{tenant_id}/... or /{tenant_id}
func extractTenantIDFromPath(path string) string {
	// Remove leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Find first slash (end of tenant_id segment)
	slashIdx := strings.Index(path, "/")
	if slashIdx == -1 {
		// No slash found, entire remaining path might be tenant_id
		if len(path) > 0 {
			return path
		}
		return ""
	}

	// Return the first segment (tenant_id)
	return path[:slashIdx]
}
