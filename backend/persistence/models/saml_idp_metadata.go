package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type SamlIDPMetadata struct {
	ID               uuid.UUID `db:"id"`
	TenantID         uuid.UUID `db:"tenant_id"`
	ProviderID       uuid.UUID `db:"provider_id"`
	RawMetadataXML   string    `db:"raw_metadata_xml"`
	Issuer           string    `db:"issuer"`
	SSOURL           string    `db:"sso_url"`
	CertificatesPEM  string    `db:"certificates_pem"`
	LastFetchedAt    time.Time `db:"last_fetched_at"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func (m SamlIDPMetadata) TableName() string {
	return "saml_idp_metadata"
}
