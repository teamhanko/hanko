package test

import (
	"context"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

func NewOidcKeysPersister(init []models.Key) persistence.OIDCKeyPersister {
	return &oidcKeysPersister{append([]models.Key{}, init...)}
}

type oidcKeysPersister struct {
	oidcKeys []models.Key
}

func (o *oidcKeysPersister) GetSigningKey(ctx context.Context) (*models.SigningKey, error) {
	var found *models.Key

	for _, data := range o.oidcKeys {
		if data.ExpiresAt.After(time.Now()) {
			if found == nil || found.ExpiresAt.After(data.ExpiresAt) {
				found = &data
			}
		}
	}

	return found.SigningKey(), nil
}

func (o *oidcKeysPersister) GetPublicKeys(ctx context.Context) ([]models.PublicKey, error) {
	var found []models.PublicKey

	for _, data := range o.oidcKeys {
		found = append(found, data.PublicKey())
	}

	return found, nil
}
