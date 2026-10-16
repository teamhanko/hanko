package certificate

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/v3/saml"
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
		if err := saml.ValidateKeyBits(p.keyBits); err != nil {
			return err
		}
	}

	return saml.ValidateValidityYears(p.validityYears)
}

func (p certificateParams) serviceParams() saml.CertificateParams {
	return saml.CertificateParams{KeyBits: p.keyBits, ValidityYears: p.validityYears, RotateKey: p.rotateKey}
}

func keyBitsUsage() string {
	return fmt.Sprintf("RSA key size in bits, one of %v", saml.AllowedKeyBits)
}

func validityYearsUsage() string {
	return fmt.Sprintf("validity of the certificate in whole years, %d to %d", saml.MinValidityYears, saml.MaxValidityYears)
}

// addCreateFlags adds the flags that customize a newly created certificate.
func addCreateFlags(cmd *cobra.Command, p *certificateParams) {
	cmd.Flags().IntVar(&p.keyBits, "key-bits", saml.DefaultKeyBits, keyBitsUsage())
	cmd.Flags().IntVar(&p.validityYears, "validity-years", saml.DefaultValidityYears, validityYearsUsage())
}
