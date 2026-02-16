package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

const (
	TenantIDContextKey = "tenant_id"
	TenantContextKey   = "tenant"
)

// Tenant creates middleware that resolves tenant from HTTP header and optionally auto-provisions new tenants.
// When multi-tenant mode is disabled, this middleware is a no-op.
// When enabled, it reads the tenant ID from the configured HTTP header (default: X-Tenant-ID).
func Tenant(cfg config.MultiTenant, persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If multi-tenant mode is disabled, skip tenant resolution
			if !cfg.Enabled {
				return next(c)
			}

			var tenantID *uuid.UUID

			// Resolve tenant from HTTP header
			headerValue := c.Request().Header.Get(cfg.TenantHeader)
			if headerValue != "" {
				id, err := uuid.FromString(headerValue)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID format (must be UUID)")
				}
				tenantID = &id
			}

			if tenantID != nil {
				// Try to get existing tenant
				tenant, err := persister.GetTenantPersister().Get(*tenantID)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to lookup tenant")
				}

				// AUTO-PROVISION: Create tenant if it doesn't exist (when enabled)
				if tenant == nil {
					if !cfg.AutoProvision {
						return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
					}

					// Create new tenant with default values
					now := time.Now().UTC()
					tenant = &models.Tenant{
						ID:        *tenantID,
						Name:      fmt.Sprintf("Tenant-%s", tenantID.String()), // Use full UUID as name to avoid collisions
						Slug:      tenantID.String(),                               // Use full UUID as slug for uniqueness
						Enabled:   true,
						CreatedAt: now,
						UpdatedAt: now,
					}

					if err := persister.GetTenantPersister().Create(*tenant); err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "failed to auto-provision tenant")
					}
				}

				// Check if tenant is enabled
				if !tenant.Enabled {
					return echo.NewHTTPError(http.StatusForbidden, "tenant is disabled")
				}

				// Set tenant context
				c.Set(TenantIDContextKey, tenantID)
				c.Set(TenantContextKey, tenant)
			} else if !cfg.AllowGlobalUsers {
				// Tenant header is required but not provided
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s header required", cfg.TenantHeader))
			}

			return next(c)
		}
	}
}

// GetTenantID retrieves the tenant ID from the Echo context.
// Returns nil if no tenant ID is set (either multi-tenant mode is disabled or request is for a global user).
func GetTenantID(c echo.Context) *uuid.UUID {
	if tenantID, ok := c.Get(TenantIDContextKey).(*uuid.UUID); ok {
		return tenantID
	}
	return nil
}

// GetTenant retrieves the full Tenant model from the Echo context.
// Returns nil if no tenant is set.
func GetTenant(c echo.Context) *models.Tenant {
	if tenant, ok := c.Get(TenantContextKey).(*models.Tenant); ok {
		return tenant
	}
	return nil
}
