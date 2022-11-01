package test

import (
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewPrimaryEmailPersister(init []models.PrimaryEmail) persistence.PrimaryEmailPersister {
	return &primaryEmailPersister{append([]models.PrimaryEmail{}, init...)}
}

type primaryEmailPersister struct {
	primaryEmails []models.PrimaryEmail
}

func (p *primaryEmailPersister) Create(primaryEmail models.PrimaryEmail) error {
	return nil
}

func (p *primaryEmailPersister) Update(primaryEmail models.PrimaryEmail) error {
	return nil
}
