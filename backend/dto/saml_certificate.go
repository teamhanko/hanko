package dto

import "time"

type SamlCertificateResponse struct {
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
}
