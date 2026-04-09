package utils

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestTenantIDFromContext(t *testing.T) {
	e := echo.New()

	t.Run("returns nil for missing tenant_id", func(t *testing.T) {
		c := e.NewContext(nil, nil)

		tenantID, err := TenantIDFromContext(c)
		require.NoError(t, err)
		require.Nil(t, tenantID)
	})

	t.Run("returns tenant id when value is a uuid pointer", func(t *testing.T) {
		c := e.NewContext(nil, nil)
		id := uuid.Must(uuid.NewV4())
		c.Set("tenant_id", &id)

		tenantID, err := TenantIDFromContext(c)
		require.NoError(t, err)
		require.NotNil(t, tenantID)
		require.Equal(t, id, *tenantID)
	})

	t.Run("returns error for unexpected type", func(t *testing.T) {
		c := e.NewContext(nil, nil)
		c.Set("tenant_id", "not-a-uuid")

		tenantID, err := TenantIDFromContext(c)
		require.Error(t, err)
		require.Nil(t, tenantID)
	})
}
