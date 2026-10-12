package handler

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/saml"
)

type SamlCertificateHandler struct {
	persister persistence.Persister
	service   *saml.CertificateService
}

func NewSamlCertificateHandler(persister persistence.Persister) *SamlCertificateHandler {
	return &SamlCertificateHandler{persister: persister, service: saml.NewCertificateService(persister)}
}

// Renew reissues a tenant's SAML certificate. By default the key pair is kept
// and only the certificate changes, so identity providers that pinned the
// previous certificate need the new one. A key size other than the current one,
// or rotate_key, generates a new key pair instead: the identity provider then
// no longer has the matching public key, and encrypted logins fail until it is
// given the new certificate.
func (h *SamlCertificateHandler) Renew(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
	}

	var request dto.RenewSamlCertificateRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}

	params, err := renewParams(request)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tenant, err := h.persister.GetTenantPersister().Get(tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get tenant: %v", err))
	}
	if tenant == nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("tenant with ID %s not found", tenantID))
	}

	tenantConfig, err := config.ParseMultitenancyTenantConfig(tenant.Config)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	cert, err := h.persister.GetSamlCertificatePersister().GetFirst(tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get SAML certificate: %v", err))
	}
	if cert == nil {
		return echo.NewHTTPError(http.StatusNotFound, "SAML certificate not found")
	}

	keyChanged, err := h.service.Renew(cert, saml.NewCertificateSubject(tenantConfig.Service.Name, tenantConfig.Saml.Endpoint), params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to renew SAML certificate: %v", err))
	}

	response, err := certificateResponse(cert)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	response.KeyChanged = keyChanged

	return c.JSON(http.StatusOK, response)
}

// renewParams validates the request and translates it to certificate parameters. A field that is not in the
// request stays at the zero value, which means the default.
func renewParams(request dto.RenewSamlCertificateRequest) (saml.CertificateParams, error) {
	params := saml.CertificateParams{RotateKey: request.RotateKey}

	if request.KeyBits != nil {
		if err := saml.ValidateKeyBits(*request.KeyBits); err != nil {
			return saml.CertificateParams{}, err
		}
		params.KeyBits = *request.KeyBits
	}

	if request.ValidityYears != nil {
		if err := saml.ValidateValidityYears(*request.ValidityYears); err != nil {
			return saml.CertificateParams{}, err
		}
		params.ValidityYears = *request.ValidityYears
	}

	return params, nil
}

func certificateResponse(cert *models.SamlCertificate) (*dto.SamlCertificateResponse, error) {
	block, _ := pem.Decode([]byte(cert.CertData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode SAML certificate: no PEM block found")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SAML certificate: %w", err)
	}

	publicKey, ok := x509Cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("SAML certificate does not have an RSA key")
	}

	return &dto.SamlCertificateResponse{
		NotBefore: x509Cert.NotBefore,
		NotAfter:  x509Cert.NotAfter,
		KeyBits:   publicKey.N.BitLen(),
	}, nil
}
