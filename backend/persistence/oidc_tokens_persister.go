package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type OIDCAccessTokensPersister interface {
	Get(ctx context.Context, uuid uuid.UUID) (*models.AccessToken, error)
	Create(ctx context.Context, accessToken models.AccessToken) error
	Delete(ctx context.Context, accessToken models.AccessToken) error
}

type oidcAccessTokensPersister struct {
	db *pop.Connection
}

func NewOIDCAccessTokensPersister(db *pop.Connection) OIDCAccessTokensPersister {
	return &oidcAccessTokensPersister{db: db}
}

func (p *oidcAccessTokensPersister) Get(ctx context.Context, uuid uuid.UUID) (*models.AccessToken, error) {
	accessToken := models.AccessToken{}
	err := p.db.WithContext(ctx).Find(&accessToken, uuid)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	return &accessToken, nil
}

func (p *oidcAccessTokensPersister) Create(ctx context.Context, accessToken models.AccessToken) error {
	vErr, err := p.db.WithContext(ctx).ValidateAndCreate(&accessToken)
	if err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("access token object validation failed: %w", vErr)
	}

	return nil
}

func (p *oidcAccessTokensPersister) Delete(ctx context.Context, accessToken models.AccessToken) error {
	err := p.db.WithContext(ctx).Destroy(&accessToken)
	if err != nil {
		return fmt.Errorf("failed to delete access token: %w", err)
	}

	return nil
}

type OIDCRefreshTokensPersister interface {
	Get(ctx context.Context, uuid uuid.UUID) (*models.RefreshToken, error)
	Create(ctx context.Context, refreshToken models.RefreshToken) error
	Delete(ctx context.Context, refreshToken models.RefreshToken) error
	TerminateSessions(ctx context.Context, clientID string, userID string) error
}

type oidcRefreshTokensPersister struct {
	db *pop.Connection
}

func NewOIDCRefreshTokensPersister(db *pop.Connection) OIDCRefreshTokensPersister {
	return &oidcRefreshTokensPersister{db: db}
}

func (p *oidcRefreshTokensPersister) Get(ctx context.Context, uuid uuid.UUID) (*models.RefreshToken, error) {
	refreshToken := models.RefreshToken{}
	err := p.db.WithContext(ctx).Find(&refreshToken, uuid)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (p *oidcRefreshTokensPersister) Create(ctx context.Context, refreshToken models.RefreshToken) error {
	vErr, err := p.db.WithContext(ctx).ValidateAndCreate(&refreshToken)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("refresh token object validation failed: %w", vErr)
	}

	return nil
}

func (p *oidcRefreshTokensPersister) Delete(ctx context.Context, refreshToken models.RefreshToken) error {
	// Slight difference: we not only need to delete the RefreshToken - we also need to delete the associated AccessToken
	err := p.db.WithContext(ctx).Destroy(&refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (p *oidcRefreshTokensPersister) TerminateSessions(ctx context.Context, clientID string, userID string) error {
	err := p.db.WithContext(ctx).RawQuery("DELETE FROM sessions WHERE client_id = ? AND user_id = ?", clientID, userID).Exec()
	if err != nil {
		return fmt.Errorf("failed to terminate sessions: %w", err)
	}

	return nil
}
