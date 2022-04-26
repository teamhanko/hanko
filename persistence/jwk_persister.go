package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/persistence/models"
)

type JwkPersister interface {
	Get(int) (*models.Jwk, error)
	GetAll() ([]models.Jwk, error)
	GetLast() (*models.Jwk, error)
	Create(models.Jwk) error
}

type jwkPersister struct {
	db *pop.Connection
}

func NewJwkPersister(db *pop.Connection) JwkPersister {
	return &jwkPersister{db: db}
}

func (p *jwkPersister) Get(id int) (*models.Jwk, error) {
	jwk := models.Jwk{}
	err := p.db.Find(&jwk, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get jwk: %w", err)
	}
	return &jwk, nil
}

func (p *jwkPersister) GetAll() ([]models.Jwk, error) {
	jwks := []models.Jwk{}
	err := p.db.All(&jwks)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get all jwks: %w", err)
	}
	return jwks, nil
}

func (p *jwkPersister) GetLast() (*models.Jwk, error) {
	jwk := models.Jwk{}
	err := p.db.Order("id asc").Last(&jwk)
	if err != nil && err == sql.ErrNoRows {
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
