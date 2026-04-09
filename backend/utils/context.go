package utils

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

// TenantIDFromContext returns the tenant ID stored in the Echo context.
//
// A nil return value is valid and indicates single-tenant mode. If the context
// contains a tenant value of an unexpected type, the function returns an error.
func TenantIDFromContext(c echo.Context) (*uuid.UUID, error) {
	v := c.Get("tenant_id")
	if v == nil {
		return nil, nil
	}

	tenantID, ok := v.(*uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("tenant_id has unexpected type %T", v)
	}

	return tenantID, nil
}
