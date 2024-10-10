package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewSessionPersister(init []models.Session) persistence.SessionPersister {
	return &sessionPersister{sessions: init}
}

type sessionPersister struct {
	sessions []models.Session
}

func (s sessionPersister) Create(session models.Session) error {
	//TODO implement me
	panic("implement me")
}

func (s sessionPersister) Get(id uuid.UUID) (*models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (s sessionPersister) Update(session models.Session) error {
	//TODO implement me
	panic("implement me")
}

func (s sessionPersister) List(userID uuid.UUID) ([]models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (s sessionPersister) ListActive(userID uuid.UUID) ([]models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (s sessionPersister) Delete(session models.Session) error {
	//TODO implement me
	panic("implement me")
}
