package saml

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

const certificateExpiryWarningWindow = 60 * 24 * time.Hour

type expiringCertificate struct {
	TenantID uuid.UUID
	NotAfter time.Time
}

// CheckSAMLCertificateExpiry logs a single line listing every tenant whose
// SAML certificate is expired or expires within the next 60 days. It reads all
// certificates with one query, so it is cheap enough to run at every startup.
func CheckSAMLCertificateExpiry(persister persistence.Persister) error {
	expiring, err := findExpiringCertificates(persister.GetConnection(), time.Now(), certificateExpiryWarningWindow)
	if err != nil {
		return err
	}
	if len(expiring) == 0 {
		return nil
	}

	entries := make([]string, len(expiring))
	for i, cert := range expiring {
		verb := "expires"
		if cert.NotAfter.Before(time.Now()) {
			verb = "expired"
		}
		entries[i] = fmt.Sprintf("%s (%s %s)", cert.TenantID, verb, cert.NotAfter.Format("2006-01-02"))
	}

	log.Printf(
		"WARNING: SAML certificate expired or expiring within %d days for %d tenant(s): %s. Renew with `hanko saml certificate renew --tenant_id=<id>`",
		int(certificateExpiryWarningWindow.Hours()/24), len(expiring), strings.Join(entries, ", "),
	)

	return nil
}

// findExpiringCertificates returns the certificates that are expired or end
// before now+window, soonest first. Certificates that can't be parsed are
// logged and skipped.
func findExpiringCertificates(db *pop.Connection, now time.Time, window time.Duration) ([]expiringCertificate, error) {
	type row struct {
		TenantID uuid.UUID `db:"tenant_id"`
		CertData string    `db:"cert_data"`
	}

	var rows []row
	if err := db.RawQuery("SELECT tenant_id, cert_data FROM saml_certificates").All(&rows); err != nil {
		return nil, fmt.Errorf("failed to query SAML certificates: %w", err)
	}

	var expiring []expiringCertificate
	for _, r := range rows {
		block, _ := pem.Decode([]byte(r.CertData))
		if block == nil {
			log.Printf("unable to check SAML certificate expiry for tenant %s: no PEM block found", r.TenantID)
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Printf("unable to check SAML certificate expiry for tenant %s: %v", r.TenantID, err)
			continue
		}

		if cert.NotAfter.Before(now.Add(window)) {
			expiring = append(expiring, expiringCertificate{TenantID: r.TenantID, NotAfter: cert.NotAfter})
		}
	}

	sort.Slice(expiring, func(i, j int) bool { return expiring[i].NotAfter.Before(expiring[j].NotAfter) })

	return expiring, nil
}
