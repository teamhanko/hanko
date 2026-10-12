package models

import "fmt"

const (
	DefaultKeyBits       = 2048
	DefaultValidityYears = 10
	MinValidityYears     = 1
	MaxValidityYears     = 10
)

// AllowedKeyBits are the RSA key sizes certificates can be created or renewed with.
var AllowedKeyBits = []int{2048, 3072, 4096}

type certificateOptions struct {
	keyBits       int // 0 means not set: the default for new certificates, the current key for renewals
	validityYears int
	rotateKey     bool
}

// CertificateOption customizes how a SAML certificate is created or renewed.
type CertificateOption func(*certificateOptions)

// WithKeyBits sets the RSA key size. On renewal, a size other than the current one generates a new key pair.
func WithKeyBits(bits int) CertificateOption {
	return func(o *certificateOptions) { o.keyBits = bits }
}

// WithValidityYears sets how many whole years the certificate is valid for.
func WithValidityYears(years int) CertificateOption {
	return func(o *certificateOptions) { o.validityYears = years }
}

// WithRotateKey makes a renewal generate a new key pair even if the key size is unchanged.
func WithRotateKey() CertificateOption {
	return func(o *certificateOptions) { o.rotateKey = true }
}

// ValidateKeyBits reports whether bits is an allowed RSA key size.
func ValidateKeyBits(bits int) error {
	for _, allowed := range AllowedKeyBits {
		if bits == allowed {
			return nil
		}
	}
	return fmt.Errorf("key size must be one of %v bits, got %d", AllowedKeyBits, bits)
}

// ValidateValidityYears reports whether years is an allowed certificate validity.
func ValidateValidityYears(years int) error {
	if years < MinValidityYears || years > MaxValidityYears {
		return fmt.Errorf("validity must be between %d and %d years, got %d", MinValidityYears, MaxValidityYears, years)
	}
	return nil
}

func newCertificateOptions(opts []CertificateOption) (certificateOptions, error) {
	options := certificateOptions{validityYears: DefaultValidityYears}
	for _, opt := range opts {
		opt(&options)
	}

	if options.keyBits != 0 {
		if err := ValidateKeyBits(options.keyBits); err != nil {
			return certificateOptions{}, err
		}
	}
	if err := ValidateValidityYears(options.validityYears); err != nil {
		return certificateOptions{}, err
	}

	return options, nil
}

func (o certificateOptions) keyBitsOrDefault() int {
	if o.keyBits == 0 {
		return DefaultKeyBits
	}
	return o.keyBits
}

// replacesKey reports whether a renewal of a certificate whose key has currentBits bits generates a new key pair.
func (o certificateOptions) replacesKey(currentBits int) bool {
	return o.rotateKey || (o.keyBits != 0 && o.keyBits != currentBits)
}
