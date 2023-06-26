package models

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"gopkg.in/square/go-jose.v2"
	"strings"
	"time"
)

type Key struct {
	ID         uuid.UUID               `db:"id" json:"id"`
	Algo       jose.SignatureAlgorithm `db:"algorithm" json:"algorithm"`
	Key        string                  `db:"public_key" json:"public_key"`
	PrivateKey string                  `db:"private_key" json:"private_key"`
	ExpiresAt  time.Time               `db:"expires_at" json:"expires_at"`
}

func (k *Key) SigningKey() *SigningKey {
	var key interface{}
	switch k.Algo {
	case jose.RS256, jose.RS384, jose.RS512:
		block, _ := pem.Decode([]byte(strings.Replace(k.PrivateKey, "\\n", "\n", -1)))
		if block == nil {
			panic("failed to parse PEM block containing the key")
		}

		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}

		key = priv
	default:
		panic("not implemented")
	}

	return &SigningKey{
		keyID:      k.ID,
		algorithm:  k.Algo,
		privateKey: key,
	}
}

func (k *Key) PublicKey() PublicKey {
	return PublicKey{
		keyID:     k.ID,
		algorithm: k.Algo,
		publicKey: k.Key,
	}
}

func (k *Key) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: k.ID},
		&validators.StringIsPresent{Name: "Algorithm", Field: string(k.Algo)},
	), nil
}

type SigningKey struct {
	keyID      uuid.UUID
	algorithm  jose.SignatureAlgorithm
	privateKey interface{}
}

func (k *SigningKey) ID() string {
	return k.keyID.String()
}

func (k *SigningKey) SignatureAlgorithm() jose.SignatureAlgorithm {
	return k.algorithm
}

func (k *SigningKey) Key() interface{} {
	return k.privateKey
}

type PublicKey struct {
	keyID     uuid.UUID
	algorithm jose.SignatureAlgorithm
	publicKey interface{}
}

func (k *PublicKey) ID() string {
	return k.keyID.String()
}

func (k *PublicKey) Key() interface{} {
	return k.publicKey
}

func (k *PublicKey) Algorithm() jose.SignatureAlgorithm {
	return k.algorithm
}

func (k *PublicKey) Use() string {
	return "sig"
}
