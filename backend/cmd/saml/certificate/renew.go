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
	"github.com/teamhanko/hanko/backend/v3/saml"
)

func NewRenewCommand() *cobra.Command {
	var (
		configFile string
		tenantID   string
		dryRun     bool
		params     certificateParams
	)

	cmd := &cobra.Command{
		Use:   "renew",
		Short: "reissue a tenant's SAML certificate",
		Long: `Reissues a tenant's SAML certificate with a new validity period.

The common name and subject alternative name of the certificate are derived from the tenant's current SAML
endpoint again, so renewing after the endpoint changed, for example because of a custom domain, gives the
certificate the new host name.

By default the key pair is kept and only the certificate changes. Identity providers that pinned the previous
certificate then need the new one, though one that only checks the public key keeps working, and encrypted logins
still work because the private key is unchanged.

A different --key-bits, or --rotate-key, generates a new key pair instead. The identity provider then no longer
has the matching public key: encrypted logins fail and signatures may be rejected until it is given the new
certificate.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if tenantID == "" {
				return errors.New("tenant_id is required")
			}
			if _, err := uuid.FromString(tenantID); err != nil {
				return errors.New("tenant_id is not a uuid")
			}

			// without --key-bits the size defaults to the current key's, which is only known once the
			// certificate is loaded, so the size is validated here only when given
			if cmd.Flags().Changed("key-bits") {
				if err := saml.ValidateKeyBits(params.keyBits); err != nil {
					return err
				}
			}

			return params.validate()
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

			certSubjectResolver := newSubjectResolver(cfg.MultiTenancy.Enabled, cfg.TenantConfig)

			result, err := renewCertificateForTenant(certSubjectResolver, dbPersister, tID, dryRun, params)
			if err != nil {
				log.Fatalf("failed to renew certificate: %s", err)
			}
			fmt.Println(result.Message)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", config.DefaultConfigFilePath, "config file")
	cmd.Flags().StringVar(&tenantID, "tenant_id", "", "tenant ID to renew the SAML certificate of")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without renewing the certificate")
	cmd.Flags().IntVar(&params.keyBits, "key-bits", 0, keyBitsUsage()+"; a size other than the current one generates a new key pair (default: the current key size)")
	cmd.Flags().IntVar(&params.validityYears, "validity-years", saml.DefaultValidityYears, validityYearsUsage())
	cmd.Flags().BoolVar(&params.rotateKey, "rotate-key", false, "generate a new key pair even if the key size stays the same")

	return cmd
}

func renewCertificateForTenant(certSubjectResolver subjectResolver, persister persistence.Persister, tenantID uuid.UUID, dryRun bool, params certificateParams) (certResult, error) {
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

	service := saml.NewCertificateService(persister)
	subject, err := certSubjectResolver.Resolve(tenant)
	if err != nil {
		return certResult{}, err
	}

	current, err := parseCertificate(cert)
	if err != nil {
		return certResult{}, err
	}
	validity := fmt.Sprintf("valid %s to %s", current.NotBefore, current.NotAfter)

	// the key size defaults to the current key's, so renewing only replaces the key if a different size
	// or a rotation is asked for
	if params.keyBits == 0 {
		if params.keyBits, err = cert.KeyBits(); err != nil {
			return certResult{}, err
		}
	}

	if dryRun {
		replacesKey, err := service.WouldReplaceKey(cert, params.serviceParams())
		if err != nil {
			return certResult{}, err
		}

		return certResult{outcomeWouldRenew, fmt.Sprintf(
			"tenant %s: would renew certificate (currently %s) for %d years, %s, key pair %s",
			tenantID, validity, params.validityYears, subject, keyOutcome(replacesKey, params.keyBits),
		)}, nil
	}

	keyChanged, err := service.Renew(cert, subject, params.serviceParams())
	if err != nil {
		return certResult{}, err
	}

	renewed, err := parseCertificate(cert)
	if err != nil {
		return certResult{}, err
	}

	warning := "identity providers that pinned the previous certificate need the new one"
	if keyChanged {
		warning = "the key pair changed: identity providers need the new certificate, until then encrypted logins fail and signatures may be rejected"
	}

	return certResult{outcomeRenewed, fmt.Sprintf(
		"renewed SAML certificate for tenant %s (was %s, now valid %s to %s)\n%s\nkey pair %s\n%s",
		tenantID, validity, renewed.NotBefore, renewed.NotAfter, subject, keyOutcome(keyChanged, params.keyBits), warning,
	)}, nil
}

func keyOutcome(replaced bool, keyBits int) string {
	if replaced {
		return fmt.Sprintf("replaced with a new %d-bit key", keyBits)
	}

	return "kept"
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
