package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/persistence/models"
)

func NewPasscodePersister(init []models.Passcode) persistence.PasscodePersister {
	return &passcodePersister{append([]models.Passcode{}, init...)}
}

type passcodePersister struct {
	passcodes []models.Passcode
}

func (p *passcodePersister) Get(id uuid.UUID) (*models.Passcode, error) {
	var found *models.Passcode
	for _, data := range p.passcodes {
		if data.ID == id {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (p *passcodePersister) Create(passcode models.Passcode) error {
	p.passcodes = append(p.passcodes, passcode)
	return nil
}

func (p *passcodePersister) Update(passcode models.Passcode) error {
	for i, data := range p.passcodes {
		if data.ID == passcode.ID {
			p.passcodes[i] = passcode
		}
	}
	return nil
}

func (p *passcodePersister) Delete(passcode models.Passcode) error {
	index := -1
	for i, data := range p.passcodes {
		if data.ID == passcode.ID {
			index = i
		}
	}
	if index > -1 {
		p.passcodes = append(p.passcodes[:index], p.passcodes[index+1:]...)
	}

	return nil
}
