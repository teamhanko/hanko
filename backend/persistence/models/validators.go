package models

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// TenantIDIsPresent validates that a tenant ID is a well-formed UUID, same as
// validators.UUIDIsPresent, except it accepts uuid.Nil: some deployments use
// the nil UUID as their tenant ID.
type TenantIDIsPresent struct {
	Name    string
	Field   uuid.UUID
	Message string
}

func (v *TenantIDIsPresent) IsValid(errors *validate.Errors) {
	if strings.TrimSpace(v.Field.String()) != "" {
		return
	}

	if len(v.Message) > 0 {
		errors.Add(validators.GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("%s can not be blank.", v.Name))
}
