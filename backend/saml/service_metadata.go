package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/russellhaering/gosaml2/types"
	dsigTypes "github.com/russellhaering/goxmldsig/types"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// ParsedMetadata contains the extracted metadata from IdP XML
type ParsedMetadata struct {
	EntityID        string
	Issuer          string
	SSOURL          string
	CertificatesPEM []string
	RawXML          string
}

// SamlMetadataService handles fetching and caching of SAML IdP metadata
type SamlMetadataService struct {
	persister persistence.Persister
}

func NewSamlMetadataService(persister persistence.Persister) *SamlMetadataService {
	return &SamlMetadataService{
		persister: persister,
	}
}

// FetchAndParse fetches IdP metadata from URL and parses it
func (s *SamlMetadataService) FetchAndParse(metadataURL string) (*ParsedMetadata, error) {
	response, err := http.Get(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch metadata: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for idp metadata failed with status code: %v", response.StatusCode)
	}

	metadataBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read idp metadata response body: %w", err)
	}

	// Parse XML metadata
	idpMetadata := &types.EntityDescriptor{}
	err = xml.Unmarshal(metadataBody, idpMetadata)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal idp metadata response body to xml: %w", err)
	}

	// Extract certificates as PEM strings
	certificatesPEM := []string{}
	for _, keyDescriptor := range idpMetadata.IDPSSODescriptor.KeyDescriptors {
		for index, x509Certificate := range keyDescriptor.KeyInfo.X509Data.X509Certificates {
			pemCert, err := extractCertificatePEM(index, x509Certificate)
			if err != nil {
				return nil, err
			}
			certificatesPEM = append(certificatesPEM, pemCert)
		}
	}

	// Extract SSO URL (prefer POST binding, fallback to first available)
	ssoURL := ""
	for _, ssoService := range idpMetadata.IDPSSODescriptor.SingleSignOnServices {
		if ssoService.Binding == "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" {
			ssoURL = ssoService.Location
			break
		}
	}
	if ssoURL == "" && len(idpMetadata.IDPSSODescriptor.SingleSignOnServices) > 0 {
		ssoURL = idpMetadata.IDPSSODescriptor.SingleSignOnServices[0].Location
	}

	return &ParsedMetadata{
		EntityID:        idpMetadata.EntityID,
		Issuer:          idpMetadata.EntityID, // EntityID typically serves as Issuer
		SSOURL:          ssoURL,
		CertificatesPEM: certificatesPEM,
		RawXML:          string(metadataBody),
	}, nil
}

// Store saves metadata to the database cache
func (s *SamlMetadataService) Store(tenantID uuid.UUID, providerID uuid.UUID, metadata *ParsedMetadata) error {
	// Convert certificates array to JSON string for storage
	certsJSON, err := json.Marshal(metadata.CertificatesPEM)
	if err != nil {
		return fmt.Errorf("failed to marshal certificates: %w", err)
	}

	metadataModel := models.SamlIDPMetadata{
		ID:              uuid.Must(uuid.NewV4()),
		TenantID:        tenantID,
		ProviderID:      providerID,
		RawMetadataXML:  metadata.RawXML,
		Issuer:          metadata.Issuer,
		SSOURL:          metadata.SSOURL,
		CertificatesPEM: string(certsJSON),
		LastFetchedAt:   time.Now(),
	}

	// Wrap check-then-act in transaction to prevent race conditions
	return s.persister.Transaction(func(tx *pop.Connection) error {
		metadataPersister := s.persister.GetSamlIDPMetadataPersisterWithConnection(tx)

		// Check if metadata already exists
		existing, err := metadataPersister.Get(tenantID, providerID)
		if err != nil {
			return fmt.Errorf("failed to check existing metadata: %w", err)
		}

		if existing != nil {
			// Update existing
			metadataModel.ID = existing.ID
			metadataModel.CreatedAt = existing.CreatedAt
			return metadataPersister.Update(metadataModel)
		}

		// Create new
		return metadataPersister.Create(metadataModel)
	})
}

// Get retrieves metadata from cache
func (s *SamlMetadataService) Get(tenantID uuid.UUID, providerID uuid.UUID) (*ParsedMetadata, error) {
	metadata, err := s.persister.GetSamlIDPMetadataPersister().Get(tenantID, providerID)
	if err != nil {
		return nil, err
	}

	if metadata == nil {
		return nil, nil
	}

	// Parse certificates JSON
	var certificatesPEM []string
	err = json.Unmarshal([]byte(metadata.CertificatesPEM), &certificatesPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal certificates: %w", err)
	}

	return &ParsedMetadata{
		EntityID:        metadata.ProviderID.String(), // ProviderID is derived from EntityID
		Issuer:          metadata.Issuer,
		SSOURL:          metadata.SSOURL,
		CertificatesPEM: certificatesPEM,
		RawXML:          metadata.RawMetadataXML,
	}, nil
}

// extractCertificatePEM extracts a certificate as PEM string
func extractCertificatePEM(index int, x509Certificate dsigTypes.X509Certificate) (string, error) {
	if x509Certificate.Data == "" {
		return "", fmt.Errorf("metadata contains an empty certificate at index %d", index)
	}

	stringifiedData := strings.TrimSpace(strings.ReplaceAll(x509Certificate.Data, "\n", ""))

	certData, err := base64.StdEncoding.DecodeString(stringifiedData)
	if err != nil {
		return "", fmt.Errorf("unable to decode certificate at index %d: %w", index, err)
	}

	// Verify it's a valid certificate
	_, err = x509.ParseCertificate(certData)
	if err != nil {
		return "", fmt.Errorf("unable to parse certificate at index %d: %w", index, err)
	}

	// Return as PEM-formatted string
	return fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", stringifiedData), nil
}
