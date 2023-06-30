package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type OIDCKeyPersister interface {
	GetSigningKey(ctx context.Context) (*models.SigningKey, error)
	GetPublicKeys(ctx context.Context) ([]models.PublicKey, error)
}

type oidcKeysPersister struct {
	db *pop.Connection
}

func NewOIDCKeyPersister(db *pop.Connection) OIDCKeyPersister {
	return &oidcKeysPersister{db: db}
}

func (p *oidcKeysPersister) GetSigningKey(ctx context.Context) (*models.SigningKey, error) {
	key := models.Key{}
	err := p.db.WithContext(ctx).Where("expires_at > ?", time.Now()).Order("expires_at asc").First(&key)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}

	return key.SigningKey(), nil
}

func (p *oidcKeysPersister) GetPublicKeys(ctx context.Context) ([]models.PublicKey, error) {
	var keys []models.Key
	err := p.db.WithContext(ctx).All(&keys)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get public keys: %w", err)
	}

	var publicKeys []models.PublicKey
	for _, key := range keys {
		publicKeys = append(publicKeys, key.PublicKey())
	}

	return publicKeys, nil
}
