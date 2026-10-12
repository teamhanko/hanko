package handler

import (
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
)

type SamlCertificateHandler struct {
	persister persistence.Persister
}

func NewSamlCertificateHandler(persister persistence.Persister) *SamlCertificateHandler {
	return &SamlCertificateHandler{persister: persister}
}

// Get returns the validity period of a tenant's SAML certificate.
func (h *SamlCertificateHandler) Get(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
	}

	cert, err := h.persister.GetSamlCertificatePersister().GetFirst(tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get SAML certificate: %v", err))
	}
	if cert == nil {
		return echo.NewHTTPError(http.StatusNotFound, "SAML certificate not found")
	}

	response, err := certificateResponse(cert)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

// Renew reissues a tenant's SAML certificate. The key pair is kept, but the
// certificate itself changes, so identity providers that pinned the previous
// certificate need the new one.
func (h *SamlCertificateHandler) Renew(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
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

	if err := h.persister.GetSamlCertificatePersister().Renew(cert, tenantConfig.Service.Name); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to renew SAML certificate: %v", err))
	}

	response, err := certificateResponse(cert)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, response)
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

	return &dto.SamlCertificateResponse{
		NotBefore: x509Cert.NotBefore,
		NotAfter:  x509Cert.NotAfter,
	}, nil
}
