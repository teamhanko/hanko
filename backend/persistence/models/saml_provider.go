package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

type SamlProvider struct {
	ID                    uuid.UUID       `db:"id" json:"id"`
	TenantID              uuid.UUID       `db:"tenant_id" json:"tenant_id"`
	Name                  string          `db:"name" json:"name"`
	EntityID              string          `db:"entity_id" json:"entity_id"`
	MetadataURL           string          `db:"metadata_url" json:"metadata_url"`
	Domain                string          `db:"domain" json:"domain"`
	Enabled               bool            `db:"enabled" json:"enabled"`
	SkipEmailVerification bool            `db:"skip_email_verification" json:"skip_email_verification"`
	AttributeMap          json.RawMessage `db:"attribute_map" json:"attribute_map"`
	CreatedAt             time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time       `db:"updated_at" json:"updated_at"`
}

func (s *SamlProvider) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
