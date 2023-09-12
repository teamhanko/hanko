package persistence

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SessionPersister interface {
	Create(session models.Session) error
	Get(id string) (*models.Session, error)
	Delete(id string) error
}

type sessionPersister struct {
	db *pop.Connection
}

func NewSessionPersister(db *pop.Connection) SessionPersister {
	return &sessionPersister{db: db}
}

func (p *sessionPersister) Create(session models.Session) error {
	vErr, err := p.db.ValidateAndCreate(&session)
	if err != nil {
		return err
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("session object validation failed: %w", vErr)
	}

	return nil
}

func (p *sessionPersister) Get(id string) (*models.Session, error) {
	session := &models.Session{}
	err := p.db.Find(session, id)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (p *sessionPersister) Delete(id string) error {
	session := &models.Session{}
	err := p.db.Find(session, id)
	if err != nil {
		return err
	}

	return p.db.Destroy(session)
}
