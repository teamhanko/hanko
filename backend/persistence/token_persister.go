package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type TokenPersister interface {
	Create(token models.Token) error
	GetByValue(value string) (*models.Token, error)
	Delete(token models.Token) error
}

type tokenPersister struct {
	db *pop.Connection
}

func NewTokenPersister(db *pop.Connection) TokenPersister {
	return &tokenPersister{db: db}
}

func (t tokenPersister) Create(token models.Token) error {
	vErr, err := t.db.ValidateAndCreate(&token)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("token object validation failed: %w", vErr)
	}

	return nil
}

func (t tokenPersister) GetByValue(value string) (*models.Token, error) {
	token := models.Token{}
	err := t.db.Where("value = ?", value).First(&token)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token by value: %w", err)
	}

	return &token, nil
}

func (t tokenPersister) Delete(token models.Token) error {
	err := t.db.Destroy(&token)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}
