package local_db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/v2/crypto/aes_gcm"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type DefaultManager struct {
	encrypter *aes_gcm.AESGCM
	persister persistence.JwkPersister
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
	// for every key we should check if a jwk with index exists and create one if not.
	for i := range keys {
		j, err := persister.Get(i + 1)
		if j == nil && err == nil {
			_, err := manager.GenerateKey()
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	}

	return manager, nil
}

// GenerateKey generates a new RSA key and persists it to the database
func (m *DefaultManager) GenerateKey() (jwk.Key, error) {
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
	}
	err = m.persister.Create(model)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetSigningKey returns the private key used for signing
func (m *DefaultManager) GetSigningKey() (jwk.Key, error) {
	sigModel, err := m.persister.GetLast()
	if err != nil {
		return nil, err
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

// GetPublicKeys returns all public keys
func (m *DefaultManager) GetPublicKeys() (jwk.Set, error) {
	modelList, err := m.persister.GetAll()
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
func (m *DefaultManager) Sign(token jwt.Token) ([]byte, error) {
	key, err := m.GetSigningKey()
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
func (m *DefaultManager) Verify(signed []byte) (jwt.Token, error) {
	keys, err := m.GetPublicKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get public keys: %w", err)
	}
	token, err := jwt.Parse(signed, jwt.WithKeySet(keys))
	if err != nil {
		return nil, fmt.Errorf("failed to verify jwt: %w", err)
	}
	return token, nil
}
