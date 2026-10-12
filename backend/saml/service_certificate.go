package saml

import (
	"fmt"
	"time"

	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// CertificateParams customizes a SAML certificate that is created or renewed.
type CertificateParams struct {
	// KeyBits is the RSA key size. Zero means the default size for a new certificate and the size of the current
	// key when renewing.
	KeyBits int
	// ValidityYears is how many whole years the certificate is valid for. Zero means DefaultValidityYears.
	ValidityYears int
	// RotateKey makes a renewal generate a new key pair even if the key size is unchanged.
	RotateKey bool
}

func (p CertificateParams) validated() (CertificateParams, error) {
	if p.KeyBits != 0 {
		if err := ValidateKeyBits(p.KeyBits); err != nil {
			return CertificateParams{}, err
		}
	}

	if p.ValidityYears == 0 {
		p.ValidityYears = DefaultValidityYears
	}
	if err := ValidateValidityYears(p.ValidityYears); err != nil {
		return CertificateParams{}, err
	}

	return p, nil
}

// replacesKey reports whether a renewal of a certificate whose key has currentBits bits generates a new key pair.
func (p CertificateParams) replacesKey(currentBits int) bool {
	return p.RotateKey || (p.KeyBits != 0 && p.KeyBits != currentBits)
}

// CertificateService creates and renews the certificates of SAML service providers.
type CertificateService struct {
	persister persistence.Persister
}

func NewCertificateService(persister persistence.Persister) *CertificateService {
	return &CertificateService{persister: persister}
}

// Generate creates a certificate with a new key pair. It is not persisted and has no tenant yet.
func (s *CertificateService) Generate(serviceName string, params CertificateParams) (*models.SamlCertificate, error) {
	params, err := params.validated()
	if err != nil {
		return nil, err
	}

	bits := params.KeyBits
	if bits == 0 {
		bits = DefaultKeyBits
	}

	key, err := GenerateKey(bits)
	if err != nil {
		return nil, err
	}

	certData, err := SelfSign(serviceName, key, params.ValidityYears, time.Now())
	if err != nil {
		return nil, err
	}

	return models.NewSamlCertificate(certData, key)
}

// WouldReplaceKey reports whether renewing cert with params generates a new key pair.
func (s *CertificateService) WouldReplaceKey(cert *models.SamlCertificate, params CertificateParams) (bool, error) {
	params, err := params.validated()
	if err != nil {
		return false, err
	}

	currentBits, err := cert.KeyBits()
	if err != nil {
		return false, err
	}

	return params.replacesKey(currentBits), nil
}

// Renew reissues the certificate with a new validity period and persists it. The key pair is kept unless a key
// size other than the current one is requested or the key is explicitly rotated, in which case a new key pair is
// generated. It reports whether the key pair was replaced.
func (s *CertificateService) Renew(cert *models.SamlCertificate, serviceName string, params CertificateParams) (bool, error) {
	keyChanged, err := renewCertificate(cert, serviceName, params)
	if err != nil {
		return false, err
	}

	if err := s.persister.GetSamlCertificatePersister().Update(cert); err != nil {
		return false, err
	}

	return keyChanged, nil
}

// renewCertificate reissues the certificate in place, without persisting it.
func renewCertificate(cert *models.SamlCertificate, serviceName string, params CertificateParams) (bool, error) {
	params, err := params.validated()
	if err != nil {
		return false, err
	}

	key, err := cert.PrivateKey()
	if err != nil {
		return false, err
	}

	currentBits := key.N.BitLen()
	keyChanged := params.replacesKey(currentBits)
	if keyChanged {
		bits := params.KeyBits
		if bits == 0 {
			bits = currentBits
		}

		key, err = GenerateKey(bits)
		if err != nil {
			return false, err
		}
	}

	certData, err := SelfSign(serviceName, key, params.ValidityYears, time.Now())
	if err != nil {
		return false, fmt.Errorf("unable to renew certificate: %w", err)
	}

	if keyChanged {
		if err := cert.SetPrivateKey(key); err != nil {
			return false, err
		}
	}
	cert.CertData = certData

	return keyChanged, nil
}
