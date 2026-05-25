package context

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
)

type Tenant struct {
	ID     uuid.UUID           `json:"id"`
	Config config.TenantConfig `json:"config"`
}
