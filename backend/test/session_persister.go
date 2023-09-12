package test

import (
	"errors"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewSessionPersister(init []models.Session) persistence.SessionPersister {
	p := &sessionPersister{tokens: make(map[string]models.Session)}
	for _, s := range init {
		p.tokens[s.ID] = s
	}

	return p
}

type sessionPersister struct {
	tokens map[string]models.Session
}

func (s sessionPersister) Create(session models.Session) error {
	s.tokens[session.ID] = session
	return nil
}

func (s sessionPersister) Get(id string) (*models.Session, error) {
	tok, ok := s.tokens[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return &tok, nil
}

func (s sessionPersister) Delete(id string) error {
	delete(s.tokens, id)

	return nil
}
