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
	"github.com/teamhanko/hanko/backend/v3/saml"
)

type certOutcome int

const (
	outcomeCreated certOutcome = iota
	outcomeWouldCreate
	outcomeAlreadyExists
	outcomeTenantNotFound
	outcomeRenewed
	outcomeWouldRenew
)

type certResult struct {
	Outcome certOutcome
	Message string
}

// createCertificateForTenant creates a SAML certificate for a single tenant,
// if it still doesn't have one.
//
// The existence check, tenant lookup, and certificate generation all happen
// without holding any lock. Generating an RSA key takes from around 100ms
// (2048 bit) to around a second (4096 bit), far too long to hold a row lock
// for. The tenant row is only locked (via FOR UPDATE) for the final, fast
// re-check-then-insert, so this is safe to call concurrently (e.g. once per
// tenant from create-all) without two calls racing to create a certificate for
// the same tenant.
func createCertificateForTenant(certSubjectResolver subjectResolver, persister persistence.Persister, tenantID uuid.UUID, dryRun bool, params certificateParams) (certResult, error) {
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

	subject := certSubjectResolver.Resolve(tenant)

	if dryRun {
		return certResult{outcomeWouldCreate, fmt.Sprintf(
			"tenant %s: would create certificate (%s, %d-bit key, valid %d years)",
			tenantID, subject, params.keyBits, params.validityYears,
		)}, nil
	}

	cert, err := saml.NewCertificateService(persister).Generate(subject, params.serviceParams())
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

		result = certResult{outcomeCreated, fmt.Sprintf(
			"created SAML certificate for tenant %s (%s, %d-bit key, valid %d years)",
			tenantID, subject, params.keyBits, params.validityYears,
		)}
		return nil
	})

	return result, err
}

// subjectResolver resolves the subject of a tenant's certificate, which is
// derived from the tenant's SAML endpoint. Where the tenant's config comes
// from depends on the deployment mode, and is why the resolver exists: without
// multitenancy it is the config file, the default tenant's row in the database
// doesn't hold it.
type subjectResolver struct {
	multiTenancy bool
	fileConfig   config.TenantConfig
}

func newSubjectResolver(multiTenancy bool, fileConfig config.TenantConfig) subjectResolver {
	return subjectResolver{multiTenancy: multiTenancy, fileConfig: fileConfig}
}

// Resolve returns the subject of the tenant's certificate. The service name is
// the common name if there is no usable endpoint. With multitenancy, the tenant
// ID takes its place if the tenant's config has no service name or can't be
// read.
func (r subjectResolver) Resolve(tenant *models.Tenant) saml.CertificateSubject {
	if !r.multiTenancy {
		return saml.NewCertificateSubject(r.fileConfig.Service.Name, r.fileConfig.Saml.Endpoint)
	}

	var tenantConfig config.TenantConfig
	if len(tenant.Config) > 0 {
		k := koanf.New(".")
		if err := k.Load(rawbytes.Provider(tenant.Config), koanfJson.Parser()); err == nil {
			_ = k.Unmarshal("", &tenantConfig)
		}
	}

	serviceName := tenantConfig.Service.Name
	if serviceName == "" {
		serviceName = tenant.ID.String()
	}

	return saml.NewCertificateSubject(serviceName, tenantConfig.Saml.Endpoint)
}
