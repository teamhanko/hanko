package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/saml"
)

type SamlProviderHandler struct {
	providerManagementService *saml.SamlProviderManagementService
	persister                 persistence.Persister
	cfg                       *config.Config
}

func NewSamlProviderHandler(cfg *config.Config, persister persistence.Persister) *SamlProviderHandler {
	return &SamlProviderHandler{
		providerManagementService: saml.NewSamlProviderManagementService(persister),
		persister:                 persister,
		cfg:                       cfg,
	}
}

// Create creates a new SAML provider for a tenant
func (h *SamlProviderHandler) Create(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
	}

	err = h.ensureSamlEnabled(tenantID)
	if err != nil {
		return err
	}

	var req dto.CreateSamlProviderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("validation failed: %v", err))
	}

	// Use empty attribute map if not provided
	attributeMap := config.AttributeMap{}
	if req.AttributeMap != nil {
		attributeMap = *req.AttributeMap
	}

	// Create provider
	provider, err := h.providerManagementService.Create(
		tenantID,
		req.Name,
		req.MetadataURL,
		req.Domain,
		req.Enabled,
		req.SkipEmailVerification,
		attributeMap,
	)
	if err != nil {
		if errors.Is(err, saml.ErrorSamlProviderAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("failed to create provider: %v", err))
		}
		if errors.Is(err, saml.ErrorSamlProviderMetadataValidation) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to create provider: %s", err.Error()))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to create provider: %v", err))
	}

	return c.JSON(http.StatusCreated, dto.FromSamlProvider(provider))
}

// List lists all SAML providers for a tenant
func (h *SamlProviderHandler) List(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
	}

	err = h.ensureSamlEnabled(tenantID)
	if err != nil {
		return err
	}

	providers, err := h.providerManagementService.List(tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to list providers: %v", err))
	}

	return c.JSON(http.StatusOK, dto.FromSamlProviders(providers))
}

// Get retrieves a specific SAML provider
func (h *SamlProviderHandler) Get(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
	}

	err = h.ensureSamlEnabled(tenantID)
	if err != nil {
		return err
	}

	providerID, err := uuid.FromString(c.Param("providerId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid provider ID")
	}

	provider, err := h.providerManagementService.Get(tenantID, providerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get provider: %v", err))
	}

	if provider == nil {
		return echo.NewHTTPError(http.StatusNotFound, "provider not found")
	}

	return c.JSON(http.StatusOK, dto.FromSamlProvider(provider))
}

// Update updates an existing SAML provider
func (h *SamlProviderHandler) Update(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
	}

	err = h.ensureSamlEnabled(tenantID)
	if err != nil {
		return err
	}

	providerID, err := uuid.FromString(c.Param("providerId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid provider ID")
	}

	var req dto.UpdateSamlProviderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("validation failed: %v", err))
	}

	// Use empty attribute map if not provided
	attributeMap := config.AttributeMap{}
	if req.AttributeMap != nil {
		attributeMap = *req.AttributeMap
	}

	// Update provider
	err = h.providerManagementService.Update(
		tenantID,
		providerID,
		req.Name,
		req.MetadataURL,
		req.Domain,
		req.Enabled,
		req.SkipEmailVerification,
		attributeMap,
	)
	if err != nil {
		if errors.Is(err, saml.ErrorSamlProviderAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("failed to create provider: %v", err))
		}
		if errors.Is(err, saml.ErrorSamlProviderMetadataValidation) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to create provider: %s", err.Error()))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to create provider: %v", err))
	}

	// Fetch updated provider
	provider, err := h.providerManagementService.Get(tenantID, providerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get updated provider: %v", err))
	}

	return c.JSON(http.StatusOK, dto.FromSamlProvider(provider))
}

// Delete deletes a SAML provider
func (h *SamlProviderHandler) Delete(c echo.Context) error {
	tenantID, err := uuid.FromString(c.Param("tenantId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID")
	}

	err = h.ensureSamlEnabled(tenantID)
	if err != nil {
		return err
	}

	providerID, err := uuid.FromString(c.Param("providerId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid provider ID")
	}

	err = h.providerManagementService.Delete(tenantID, providerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to delete provider: %v", err))
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *SamlProviderHandler) ensureSamlEnabled(tenantID uuid.UUID) error {
	tenant, err := h.persister.GetTenantPersister().Get(tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get tenant persister: %v", err))
	}

	if tenant == nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("tenant with ID %s not found", tenantID))
	}

	var tenantConfig config.TenantConfig
	err = json.Unmarshal(tenant.Config, &tenantConfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to unmarshal tenant config: %v", err))
	}

	if !tenantConfig.Saml.Enabled {
		return echo.NewHTTPError(http.StatusForbidden, "SAML is not enabled for this tenant")
	}

	return nil
}
