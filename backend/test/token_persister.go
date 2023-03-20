package test

import (
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewTokenPersister(init []models.Token) persistence.TokenPersister {
	return &tokenPersister{append([]models.Token{}, init...)}
}

type tokenPersister struct {
	tokens []models.Token
}

func (t tokenPersister) Create(token models.Token) error {
	t.tokens = append(t.tokens, token)
	return nil
}

func (t tokenPersister) GetByValue(value string) (*models.Token, error) {
	var found *models.Token
	for _, token := range t.tokens {
		if token.Value == value {
			found = &token
		}
	}
	return found, nil
}

func (t tokenPersister) Delete(token models.Token) error {
	index := -1
	for i, t := range t.tokens {
		if t.ID == token.ID {
			index = i
		}
	}
	if index > -1 {
		t.tokens = append(t.tokens[:index], t.tokens[index+1:]...)
	}

	return nil
}
