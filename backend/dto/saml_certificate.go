package dto

import "time"

// RenewSamlCertificateRequest customizes the renewal of a SAML certificate. All fields are optional.
type RenewSamlCertificateRequest struct {
	// KeyBits is the RSA key size. Without it the current key is kept; a size other than the current one
	// generates a new key pair.
	KeyBits *int `json:"key_bits"`
	// ValidityYears is the validity of the renewed certificate in whole years, defaults to 10.
	ValidityYears *int `json:"validity_years"`
	// RotateKey generates a new key pair even if the key size stays the same.
	RotateKey bool `json:"rotate_key"`
}

type SamlCertificateResponse struct {
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	KeyBits   int       `json:"key_bits"`
	// KeyChanged reports whether a new key pair was generated, in which case identity providers need the new
	// certificate before encrypted logins work again.
	KeyChanged bool `json:"key_changed"`
}
