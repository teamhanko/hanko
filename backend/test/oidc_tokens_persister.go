package test

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewOidcAccessTokensPersister(init []models.AccessToken) persistence.OIDCAccessTokenPersister {
	return &oidcAccessTokensPersister{append([]models.AccessToken{}, init...)}
}

type oidcAccessTokensPersister struct {
	oidcAccessTokens []models.AccessToken
}

func (o *oidcAccessTokensPersister) Get(ctx context.Context, uuid uuid.UUID) (*models.AccessToken, error) {
	var found *models.AccessToken

	for _, data := range o.oidcAccessTokens {
		if data.ID == uuid {
			d := data
			found = &d
		}
	}

	return found, nil
}

func (o *oidcAccessTokensPersister) Create(ctx context.Context, accessToken models.AccessToken) error {
	o.oidcAccessTokens = append(o.oidcAccessTokens, accessToken)

	return nil
}

func (o *oidcAccessTokensPersister) Delete(ctx context.Context, accessToken models.AccessToken) error {
	index := -1

	for i, data := range o.oidcAccessTokens {
		if data.ID == accessToken.ID {
			index = i
		}
	}

	if index > -1 {
		o.oidcAccessTokens = append(o.oidcAccessTokens[:index], o.oidcAccessTokens[index+1:]...)
	}

	return nil
}

func NewOidcRefreshTokensPersister(init []models.RefreshToken) persistence.OIDCRefreshTokenPersister {
	return &oidcRefreshTokensPersister{append([]models.RefreshToken{}, init...)}
}

type oidcRefreshTokensPersister struct {
	oidcRefreshTokens []models.RefreshToken
}

func (o *oidcRefreshTokensPersister) Get(ctx context.Context, uuid uuid.UUID) (*models.RefreshToken, error) {
	var found *models.RefreshToken

	for _, data := range o.oidcRefreshTokens {
		if data.ID == uuid {
			d := data
			found = &d
		}
	}

	return found, nil
}

func (o *oidcRefreshTokensPersister) Create(ctx context.Context, refreshToken models.RefreshToken) error {
	o.oidcRefreshTokens = append(o.oidcRefreshTokens, refreshToken)

	return nil
}

func (o *oidcRefreshTokensPersister) Delete(ctx context.Context, refreshToken models.RefreshToken) error {
	index := -1
	for i, data := range o.oidcRefreshTokens {
		if data.ID == refreshToken.ID {
			index = i
		}
	}

	if index > -1 {
		o.oidcRefreshTokens = append(o.oidcRefreshTokens[:index], o.oidcRefreshTokens[index+1:]...)
	}

	return nil
}

func (o *oidcRefreshTokensPersister) TerminateSessions(ctx context.Context, clientID string, userID string) error {
	for _, data := range o.oidcRefreshTokens {
		if data.ClientID == clientID && data.UserID == userID {
			_ = o.Delete(ctx, data)
		}
	}

	return nil
}
