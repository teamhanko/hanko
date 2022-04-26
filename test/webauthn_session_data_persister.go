package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/persistence/models"
)

func NewWebauthnSessionDataPersister(init []models.WebauthnSessionData) persistence.WebauthnSessionDataPersister {
	return &webauthnSessionDataPersister{init}
}

type webauthnSessionDataPersister struct {
	sessionData []models.WebauthnSessionData
}

func (p *webauthnSessionDataPersister) Get(id uuid.UUID) (*models.WebauthnSessionData, error) {
	var found *models.WebauthnSessionData
	for _, data := range p.sessionData {
		if data.ID == id {
			found = &data
		}
	}
	return found, nil
}

func (p *webauthnSessionDataPersister) GetByChallenge(challenge string) (*models.WebauthnSessionData, error) {
	var found *models.WebauthnSessionData
	for _, data := range p.sessionData {
		if data.Challenge == challenge {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (p *webauthnSessionDataPersister) Create(sessionData models.WebauthnSessionData) error {
	p.sessionData = append(p.sessionData, sessionData)
	return nil
}

func (p *webauthnSessionDataPersister) Update(sessionData models.WebauthnSessionData) error {
	for i, data := range p.sessionData {
		if data.ID == sessionData.ID {
			p.sessionData[i] = sessionData
		}
	}
	return nil
}

func (p *webauthnSessionDataPersister) Delete(sessionData models.WebauthnSessionData) error {
	index := -1
	for i, data := range p.sessionData {
		if data.ID == sessionData.ID {
			index = i
		}
	}
	if index > -1 {
		p.sessionData = append(p.sessionData[:index], p.sessionData[index+1:]...)
	}

	return nil
}
