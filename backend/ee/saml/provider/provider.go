package provider

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
	dsigTypes "github.com/russellhaering/goxmldsig/types"
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"io"
	"net/http"
	"strings"
)

type IdpMetadata struct {
	SingleSignOnUrl string
	Issuer          string
	certs           dsig.MemoryX509CertificateStore
}

type ServiceProvider interface {
	GetUserData(assertion *saml2.AssertionInfo) *thirdparty.UserData
	ProvideMetadataAsXml() ([]byte, error)
	UseDefaultAttributesIfEmpty()
	GetDomain() string
	GetService() *saml2.SAMLServiceProvider
	GetConfig() samlConfig.IdentityProvider
}

func loadCertificate(cfg *config.Config, persister persistence.SamlCertificatePersister) (dsig.X509KeyStore, error) {
	cert, err := persister.GetFirst()
	if err != nil {
		return nil, err
	}

	if cert == nil {
		cert, err = models.NewSamlCertificate(cfg.Service.Name)
		if err != nil {
			return nil, err
		}

		err = persister.Create(cert)
		if err != nil {
			return nil, err
		}
	}
	privateKey, err := cert.DecryptCertKey()
	if err != nil {
		return nil, err
	}

	keys, err := tls.X509KeyPair([]byte(cert.CertData), privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create key pair: %w", err)
	}

	keys.Leaf, err = x509.ParseCertificate(keys.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("unable to parse certificate: %w", err)
	}

	return dsig.TLSCertKeyStore(keys), nil
}

func fetchIdpMetadata(idpConfig samlConfig.IdentityProvider) (*IdpMetadata, error) {
	response, err := http.Get(idpConfig.MetadataUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch metadata: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for idp metadata failed with status code: %v", response.StatusCode)
	}

	metadataBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read idp metadata response body: %w", err)
	}

	idpMetadata := &types.EntityDescriptor{}
	err = xml.Unmarshal(metadataBody, idpMetadata)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal idp metadata response body to xml: %w", err)
	}

	idPCertStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}

	for _, keyDescriptor := range idpMetadata.IDPSSODescriptor.KeyDescriptors {
		for index, x509Certificate := range keyDescriptor.KeyInfo.X509Data.X509Certificates {
			parsedCert, err := parseCertificate(index, x509Certificate)
			if err != nil {
				return nil, err
			}
			idPCertStore.Roots = append(idPCertStore.Roots, parsedCert)
		}
	}

	return &IdpMetadata{
		SingleSignOnUrl: idpMetadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
		Issuer:          idpMetadata.EntityID,
		certs:           idPCertStore,
	}, nil
}

func parseCertificate(index int, x509Certificate dsigTypes.X509Certificate) (*x509.Certificate, error) {
	if x509Certificate.Data == "" {
		return nil, fmt.Errorf("metadata contains an empty certificate at index %d", index)
	}

	stringifiedData := strings.TrimSpace(strings.ReplaceAll(x509Certificate.Data, "\n", ""))

	certData, err := base64.StdEncoding.DecodeString(stringifiedData)
	if err != nil {
		return nil, fmt.Errorf("unable to decode certificate at index %d: %w", index, err)
	}

	idpCertificate, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, fmt.Errorf("unable to parse certificate at index %d: %w", index, err)
	}

	return idpCertificate, nil
}

func GetProvider(providerName string, cfg *config.Config, idpConfig samlConfig.IdentityProvider, persister persistence.SamlCertificatePersister) (ServiceProvider, error) {
	if strings.Contains(strings.ToLower(providerName), "auth0") {
		return NewAuth0ServiceProvider(cfg, idpConfig, persister)
	}

	return NewBaseSamlProvider(cfg, idpConfig, persister, true)
}
