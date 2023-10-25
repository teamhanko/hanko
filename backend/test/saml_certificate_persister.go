package test

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

func NewSamlCertificatePersister(init []*models.SamlCertificate) persistence.SamlCertificatePersister {
	return &samlCertificatePersister{append([]*models.SamlCertificate{}, init...)}
}

type samlCertificatePersister struct {
	samlCertificates []*models.SamlCertificate
}

func (s samlCertificatePersister) Renew(cert *models.SamlCertificate, serviceName string) error {
	key, err := cert.DecryptCertKey()
	if key == nil || err != nil {
		return fmt.Errorf("unable to decrypt private key: %w", err)
	}

	decodedKey, _ := pem.Decode(key)
	if decodedKey == nil || decodedKey.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("unable to decode private key")
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %w", err)
	}

	now := time.Now()

	newCert, err := models.GenerateCertificate(serviceName, parsedKey, now)
	if err != nil {
		return fmt.Errorf("unable to renew certificate: %w", err)
	}

	cert.CertData = newCert

	return nil
}

func (s samlCertificatePersister) Create(cert *models.SamlCertificate) error {
	s.samlCertificates = append(s.samlCertificates, cert)

	return nil
}

func (s samlCertificatePersister) GetFirst() (*models.SamlCertificate, error) {
	for i, cert := range s.samlCertificates {
		if i == 0 {
			return cert, nil
		}

	}

	return nil, errors.New("failed to get first cert")
}

func (s samlCertificatePersister) Delete(cert *models.SamlCertificate) error {
	index := -1
	for i, existingCertificate := range s.samlCertificates {
		if existingCertificate.ID == cert.ID {
			index = i
		}
	}
	if index > -1 {
		s.samlCertificates = append(s.samlCertificates[:index], s.samlCertificates[index+1:]...)
	}

	return nil
}
