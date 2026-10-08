package certificate

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	koanfJson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type certOutcome int

const (
	outcomeCreated certOutcome = iota
	outcomeWouldCreate
	outcomeAlreadyExists
	outcomeTenantNotFound
)

type certResult struct {
	Outcome certOutcome
	Message string
}

// createCertificateForTenant creates a SAML certificate for a single tenant,
// if it still doesn't have one.
//
// The existence check, tenant lookup, and certificate generation all happen
// without holding any lock. Generating an RSA-2048 key takes on the order of
// 80ms, far too long to hold a row lock for. The tenant row is only locked
// (via FOR UPDATE) for the final, fast re-check-then-insert, so this is safe
// to call concurrently (e.g. once per tenant from create-all) without two
// calls racing to create a certificate for the same tenant.
func createCertificateForTenant(persister persistence.Persister, tenantID uuid.UUID, dryRun bool) (certResult, error) {
	existing, err := persister.GetSamlCertificatePersister().GetFirst(tenantID)
	if err != nil {
		return certResult{}, fmt.Errorf("failed to check for an existing certificate: %w", err)
	}
	if existing != nil {
		return certResult{outcomeAlreadyExists, fmt.Sprintf("tenant %s already has a SAML certificate, nothing to do", tenantID)}, nil
	}

	tenant, err := persister.GetTenantPersister().Get(tenantID)
	if err != nil {
		return certResult{}, fmt.Errorf("failed to load tenant: %w", err)
	}
	if tenant == nil {
		return certResult{outcomeTenantNotFound, fmt.Sprintf("tenant %s not found", tenantID)}, nil
	}

	serviceName := tenantID.String()
	if len(tenant.Config) > 0 {
		k := koanf.New(".")
		var tenantConfig config.TenantConfig
		if err := k.Load(rawbytes.Provider(tenant.Config), koanfJson.Parser()); err == nil {
			_ = k.Unmarshal("", &tenantConfig)
		}
		if tenantConfig.Service.Name != "" {
			serviceName = tenantConfig.Service.Name
		}
	}

	if dryRun {
		return certResult{outcomeWouldCreate, fmt.Sprintf("tenant %s: would create certificate (service name: %q)", tenantID, serviceName)}, nil
	}

	cert, err := models.NewSamlCertificate(serviceName)
	if err != nil {
		return certResult{}, fmt.Errorf("failed to generate certificate: %w", err)
	}
	cert.TenantID = tenantID

	var result certResult
	err = persister.Transaction(func(tx *pop.Connection) error {
		if err := tx.RawQuery("SELECT id FROM tenants WHERE id = ? FOR UPDATE", tenantID).Exec(); err != nil {
			return fmt.Errorf("failed to lock tenant row: %w", err)
		}

		existing, err := persister.GetSamlCertificatePersisterWithConnection(tx).GetFirst(tenantID)
		if err != nil {
			return fmt.Errorf("failed to re-check for an existing certificate: %w", err)
		}
		if existing != nil {
			result = certResult{outcomeAlreadyExists, fmt.Sprintf("tenant %s already has a SAML certificate, nothing to do", tenantID)}
			return nil
		}

		if err := persister.GetSamlCertificatePersisterWithConnection(tx).Create(cert); err != nil {
			return fmt.Errorf("failed to persist certificate: %w", err)
		}

		result = certResult{outcomeCreated, fmt.Sprintf("created SAML certificate for tenant %s", tenantID)}
		return nil
	})

	return result, err
}
