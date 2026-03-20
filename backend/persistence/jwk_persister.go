package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type JwkPersister interface {
	Get(id int, tenantID *uuid.UUID) (*models.Jwk, error)
	GetAll(tenantID *uuid.UUID) ([]models.Jwk, error)
	GetLast(tenantID *uuid.UUID) (*models.Jwk, error)
	Create(models.Jwk) error
}

type jwkPersister struct {
	db *pop.Connection
}

func NewJwkPersister(db *pop.Connection) JwkPersister {
	return &jwkPersister{db: db}
}

func (p *jwkPersister) Get(id int, tenantID *uuid.UUID) (*models.Jwk, error) {
	jwk := models.Jwk{}
	query := p.db.Q()
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.Find(&jwk, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get jwk: %w", err)
	}
	return &jwk, nil
}

func (p *jwkPersister) GetAll(tenantID *uuid.UUID) ([]models.Jwk, error) {
	jwks := []models.Jwk{}
	query := p.db.Q()
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.All(&jwks)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get all jwks: %w", err)
	}
	return jwks, nil
}

func (p *jwkPersister) GetLast(tenantID *uuid.UUID) (*models.Jwk, error) {
	jwk := models.Jwk{}
	query := p.db.Q()
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}
	err := query.Last(&jwk)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get jwk: %w", err)
	}
	return &jwk, nil
}

func (p *jwkPersister) Create(jwk models.Jwk) error {
	vErr, err := p.db.ValidateAndCreate(&jwk)
	if err != nil {
		return fmt.Errorf("failed to store jwk: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("jwk object validation failed: %w", vErr)
	}

	return nil
}
