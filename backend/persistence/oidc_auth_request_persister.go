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

type OIDCAuthRequestPersister interface {
	Get(ctx context.Context, uuid uuid.UUID) (*models.AuthRequest, error)
	Create(ctx context.Context, authRequest models.AuthRequest) error
	Delete(ctx context.Context, uuid uuid.UUID) error

	StoreAuthCode(ctx context.Context, ID uuid.UUID, code string) error
	GetAuthRequestByCode(ctx context.Context, code string) (*models.AuthRequest, error)
}

type oidcAuthRequestPersister struct {
	db *pop.Connection
}

func NewOIDCAuthRequestPersister(db *pop.Connection) OIDCAuthRequestPersister {
	return &oidcAuthRequestPersister{db: db}
}

func (p *oidcAuthRequestPersister) Get(ctx context.Context, uuid uuid.UUID) (*models.AuthRequest, error) {
	authRequest := models.AuthRequest{}
	err := p.db.WithContext(ctx).Find(&authRequest, uuid)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth request: %w", err)
	}

	return &authRequest, nil
}

func (p *oidcAuthRequestPersister) Create(ctx context.Context, authRequest models.AuthRequest) error {
	vErr, err := p.db.WithContext(ctx).ValidateAndCreate(&authRequest)
	if err != nil {
		return fmt.Errorf("failed to store auth request: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("auth request object validation failed: %w", vErr)
	}

	return nil
}

func (p *oidcAuthRequestPersister) Delete(ctx context.Context, uuid uuid.UUID) error {
	err := p.db.WithContext(ctx).Destroy(&models.AuthRequest{ID: uuid})
	if err != nil {
		return fmt.Errorf("failed to delete auth request: %w", err)
	}

	return nil
}

func (p *oidcAuthRequestPersister) StoreAuthCode(ctx context.Context, ID uuid.UUID, code string) error {
	mCode := models.AuthCode{
		ID: code,
		AuthRequest: &models.AuthRequest{
			ID: ID,
		},
	}

	vErr, err := p.db.WithContext(ctx).ValidateAndCreate(&mCode)
	if err != nil {
		return fmt.Errorf("failed to store auth code: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("auth code object validation failed: %w", vErr)
	}

	return nil
}

func (p *oidcAuthRequestPersister) GetAuthRequestByCode(ctx context.Context, code string) (*models.AuthRequest, error) {
	authCode := models.AuthCode{}

	err := p.db.WithContext(ctx).EagerPreload().Find(&authCode, code)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get auth code: %w", err)
	}

	return authCode.AuthRequest, nil
}
