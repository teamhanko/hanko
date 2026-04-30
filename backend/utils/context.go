package utils

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
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
