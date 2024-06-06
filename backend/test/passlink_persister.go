package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewPasslinkPersister(init []models.Passlink) persistence.PasslinkPersister {
	return &passlinkPersister{append([]models.Passlink{}, init...)}
}

type passlinkPersister struct {
	passlinks []models.Passlink
}

func (p *passlinkPersister) Get(id uuid.UUID) (*models.Passlink, error) {
	var found *models.Passlink
	for _, data := range p.passlinks {
		if data.ID == id {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (p *passlinkPersister) Create(passlink models.Passlink) error {
	p.passlinks = append(p.passlinks, passlink)
	return nil
}

func (p *passlinkPersister) Update(passlink models.Passlink) error {
	for i, data := range p.passlinks {
		if data.ID == passlink.ID {
			p.passlinks[i] = passlink
		}
	}
	return nil
}

func (p *passlinkPersister) Delete(passlink models.Passlink) error {
	index := -1
	for i, data := range p.passlinks {
		if data.ID == passlink.ID {
			index = i
		}
	}
	if index > -1 {
		p.passlinks = append(p.passlinks[:index], p.passlinks[index+1:]...)
	}

	return nil
}
