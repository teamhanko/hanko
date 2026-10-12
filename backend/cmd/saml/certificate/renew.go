package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func NewRenewCommand() *cobra.Command {
	var (
		configFile string
		tenantID   string
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "renew",
		Short: "reissue a tenant's SAML certificate",
		Long: `Reissues a tenant's SAML certificate with a new validity period. The key pair is kept, but the
certificate itself changes, so identity providers that pinned the previous certificate need the new one.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if tenantID == "" {
				return errors.New("tenant_id is required")
			}
			if _, err := uuid.FromString(tenantID); err != nil {
				return errors.New("tenant_id is not a uuid")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}

			tID, err := uuid.FromString(tenantID)
			if err != nil {
				log.Fatalf("invalid tenant_id: %s", err)
			}

			dbConnection, err := persistence.NewConnection(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			dbPersister := persistence.New(dbConnection)

			result, err := renewCertificateForTenant(dbPersister, tID, dryRun)
			if err != nil {
				log.Fatalf("failed to renew certificate: %s", err)
			}
			fmt.Println(result.Message)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVar(&tenantID, "tenant_id", "", "tenant ID to renew the SAML certificate of")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report the certificate's current validity without renewing it")

	return cmd
}

func renewCertificateForTenant(persister persistence.Persister, tenantID uuid.UUID, dryRun bool) (certResult, error) {
	tenant, err := persister.GetTenantPersister().Get(tenantID)
	if err != nil {
		return certResult{}, fmt.Errorf("failed to load tenant: %w", err)
	}
	if tenant == nil {
		return certResult{outcomeTenantNotFound, fmt.Sprintf("tenant %s not found", tenantID)}, nil
	}

	cert, err := persister.GetSamlCertificatePersister().GetFirst(tenantID)
	if err != nil {
		return certResult{}, fmt.Errorf("failed to load certificate: %w", err)
	}
	if cert == nil {
		return certResult{}, fmt.Errorf("tenant %s has no SAML certificate, create one with `saml certificate create`", tenantID)
	}

	current, err := parseCertificate(cert)
	if err != nil {
		return certResult{}, err
	}
	validity := fmt.Sprintf("valid %s to %s", current.NotBefore, current.NotAfter)

	if dryRun {
		return certResult{outcomeWouldRenew, fmt.Sprintf("tenant %s: would renew certificate (currently %s)", tenantID, validity)}, nil
	}

	if err := persister.GetSamlCertificatePersister().Renew(cert, tenantServiceName(tenant)); err != nil {
		return certResult{}, err
	}

	renewed, err := parseCertificate(cert)
	if err != nil {
		return certResult{}, err
	}

	return certResult{outcomeRenewed, fmt.Sprintf(
		"renewed SAML certificate for tenant %s (was %s, now valid %s to %s)\nidentity providers that pinned the previous certificate need the new one",
		tenantID, validity, renewed.NotBefore, renewed.NotAfter,
	)}, nil
}

func parseCertificate(cert *models.SamlCertificate) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(cert.CertData))
	if block == nil {
		return nil, errors.New("failed to decode certificate: no PEM block found")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return x509Cert, nil
}
