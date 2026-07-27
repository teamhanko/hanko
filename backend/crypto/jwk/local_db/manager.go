package local_db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/v3/crypto/aes_gcm"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type DefaultManager struct {
	encrypter            *aes_gcm.AESGCM
	persister            persistence.JwkPersister
	encryptionKeyVersion string
}

// NewDefaultManager creates a DefaultManager that reads and persists private keys to the database and generates new private keys when a new secret is added to the config.
// It manages the lifecycle of JSON Web Keys, handling encryption, persistence and retrieval.
func NewDefaultManager(keys []string, persister persistence.JwkPersister) (*DefaultManager, error) {
	encrypter, err := aes_gcm.NewAESGCM(keys)
	if err != nil {
		return nil, err
	}
	manager := &DefaultManager{
		encrypter: encrypter,
		persister: persister,
	}

	return manager, nil
}

// GenerateKey generates a new RSA key and persists it to the database
func (m *DefaultManager) GenerateKey(tenantID uuid.UUID) (jwk.Key, error) {
	rsa := &RSAKeyGenerator{}
	id, _ := uuid.NewV4()
	key, err := rsa.Generate(id.String())
	if err != nil {
		return nil, err
	}
	marshalled, err := json.Marshal(key)
	if err != nil {
		return nil, err
	}
	encryptedKey, err := m.encrypter.Encrypt(marshalled)
	if err != nil {
		return nil, err
	}
	model := models.Jwk{
		KeyData:   encryptedKey,
		CreatedAt: time.Now(),
		TenantId:  tenantID,
	}
	err = m.persister.Create(model)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetSigningKey returns the active private key used for signing
func (m *DefaultManager) GetSigningKey(tenantID uuid.UUID) (jwk.Key, error) {
	sigModel, err := m.persister.GetLast(tenantID)
	if err != nil {
		return nil, err
	}
	if sigModel == nil {
		return nil, fmt.Errorf("no active signing key found")
	}
	k, err := m.encrypter.Decrypt(sigModel.KeyData)
	if err != nil {
		return nil, err
	}

	key, err := jwk.ParseKey(k)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetPublicKeys returns all public keys that should be used for verification (active + rotating)
func (m *DefaultManager) GetPublicKeys(tenantID uuid.UUID) (jwk.Set, error) {
	modelList, err := m.persister.GetAll(tenantID)
	if err != nil {
		return nil, err
	}

	publicKeys := jwk.NewSet()
	for _, model := range modelList {
		k, err := m.encrypter.Decrypt(model.KeyData)
		if err != nil {
			return nil, err
		}

		key, err := jwk.ParseKey(k)

		if err != nil {
			return nil, err
		}

		publicKey, err := jwk.PublicKeyOf(key)
		if err != nil {
			return nil, err
		}
		err = publicKeys.AddKey(publicKey)
		if err != nil {
			return nil, err
		}
	}

	return publicKeys, nil
}

// Sign a JWT with the signing key and returns it
func (m *DefaultManager) Sign(token jwt.Token, tenantID uuid.UUID) ([]byte, error) {
	key, err := m.GetSigningKey(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt: %w", err)
	}
	return signed, nil
}

// Verify verifies a JWT, using the verificationKeys and returns the parsed JWT
func (m *DefaultManager) Verify(signed []byte, tenantID uuid.UUID) (jwt.Token, error) {
	keys, err := m.GetPublicKeys(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public keys: %w", err)
	}
	token, err := jwt.Parse(signed, jwt.WithKeySet(keys))
	if err != nil {
		return nil, fmt.Errorf("failed to verify jwt: %w", err)
	}
	return token, nil
}
