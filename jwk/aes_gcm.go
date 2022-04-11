package jwk

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/config"
)

// AESGCM is used to en-/decrypt the generated jwks with AES-GCM
type AESGCM struct {
	keys []string
}

func NewAESGCM(secrets config.Secrets) (*AESGCM,error) {
	if len(secrets.Keys) < 1 {
		return nil, errors.New("At least one encryption key must be provided.")
	}
	for i,v := range secrets.Keys {
		if len(v) != 16 {
			return nil, errors.New(fmt.Sprintf("Key Nr. %v has the wrong length. Is %v needs to be 16.", i, len(v)))
		}
	}
	return &AESGCM{keys: secrets.Keys}, nil
}

func (a *AESGCM) Encrypt(data []byte) (string, error) {

	return "", nil
}

func (a *AESGCM) Decrypt(data string) ([]byte, error) {

}
