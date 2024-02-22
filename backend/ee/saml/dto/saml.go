package dto

type SamlRequest struct {
	Domain string `query:"domain" validate:"required,fqdn"`
}

type SamlMetadataRequest struct {
	SamlRequest
	CertOnly bool `query:"cert_only" validate:"boolean"`
}

type SamlAuthRequest struct {
	SamlRequest
	RedirectTo string `query:"redirect_to" validate:"required,url"`
}
