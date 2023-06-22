package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"gopkg.in/square/go-jose.v2"
	"time"
)

type Key struct {
	ID         uuid.UUID               `db:"id" json:"id"`
	Algo       jose.SignatureAlgorithm `db:"algorithm" json:"algorithm"`
	Key        interface{}             `db:"public_key" json:"public_key"`
	PrivateKey interface{}             `db:"private_key" json:"private_key"`
	Expiration time.Time               `db:"expiration" json:"expiration"`
}

func (k *Key) SigningKey() *SigningKey {
	return &SigningKey{
		keyID:      k.ID,
		algorithm:  k.Algo,
		privateKey: k.PrivateKey,
	}
}

func (k *Key) PublicKey() PublicKey {
	return PublicKey{
		keyID:     k.ID,
		algorithm: k.Algo,
		publicKey: k.Key,
	}
}

func (t *Key) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: t.ID},
		&validators.StringIsPresent{Name: "Algorithm", Field: string(t.Algo)},
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
