package certificate

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// certificateParams holds the flag values that customize a certificate.
type certificateParams struct {
	// keyBits is the RSA key size. 0 means not set: the default size when creating,
	// the current key when renewing.
	keyBits       int
	validityYears int
	rotateKey     bool
}

func (p certificateParams) validate() error {
	if p.keyBits != 0 {
		if err := models.ValidateKeyBits(p.keyBits); err != nil {
			return err
		}
	}

	return models.ValidateValidityYears(p.validityYears)
}

func (p certificateParams) options() []models.CertificateOption {
	opts := []models.CertificateOption{models.WithValidityYears(p.validityYears)}
	if p.keyBits != 0 {
		opts = append(opts, models.WithKeyBits(p.keyBits))
	}
	if p.rotateKey {
		opts = append(opts, models.WithRotateKey())
	}

	return opts
}

func keyBitsUsage() string {
	return fmt.Sprintf("RSA key size in bits, one of %v", models.AllowedKeyBits)
}

func validityYearsUsage() string {
	return fmt.Sprintf("validity of the certificate in whole years, %d to %d", models.MinValidityYears, models.MaxValidityYears)
}

// addCreateFlags adds the flags that customize a newly created certificate.
func addCreateFlags(cmd *cobra.Command, p *certificateParams) {
	cmd.Flags().IntVar(&p.keyBits, "key-bits", models.DefaultKeyBits, keyBitsUsage())
	cmd.Flags().IntVar(&p.validityYears, "validity-years", models.DefaultValidityYears, validityYearsUsage())
}
