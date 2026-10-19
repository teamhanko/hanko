package models

import (
	"testing"

	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTenantIDIsPresent(t *testing.T) {
	t.Run("nil UUID is valid", func(t *testing.T) {
		v := TenantIDIsPresent{Name: "TenantID", Field: uuid.Nil}
		errors := validate.NewErrors()
		v.IsValid(errors)
		assert.False(t, errors.HasAny())
	})

	t.Run("non-nil UUID is valid", func(t *testing.T) {
		id, err := uuid.NewV4()
		assert.NoError(t, err)

		v := TenantIDIsPresent{Name: "TenantID", Field: id}
		errors := validate.NewErrors()
		v.IsValid(errors)
		assert.False(t, errors.HasAny())
	})
}
