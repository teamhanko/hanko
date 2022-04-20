package jwk

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/teamhanko/hanko/crypto/aes_gcm"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/persistence/models"
	"time"
)

type Manager interface {
	GenerateKeySet() (*jwk.KeyPair, error)
	GetKeySet(id string) *jwk.KeyPair
	GetPublicKeys() []jwk.Key
	GetSigningKey() jwk.Key
}

type DefaultManager struct {
	encrypter   *aes_gcm.AESGCM
	persister   *persistence.JwkPersister
}

func NewDefaultManager(keys []string, persister *persistence.JwkPersister) (*DefaultManager, error) {
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
			_, err := manager.GenerateKeySet()
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	}

	return manager, nil
}

func (m *DefaultManager) GenerateKeySet() (jwk.Key, error) {
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

func (m *DefaultManager) GetSigningKey() (jwk.Key, error) {
	sigModel, err := m.persister.GetLast()
	if err != nil {
		return nil, err
	}
	k, err := m.encrypter.Decrypt(sigModel.KeyData)
	if err != nil{
		return nil, err
	}
	key, err := jwk.FromRaw(k);
	if err != nil {
		return nil, err
	}
	return key, nil
}
