package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewWebauthnCredentialPersister(init []models.WebauthnCredential) persistence.WebauthnCredentialPersister {
	return &webauthnCredentialPersister{append([]models.WebauthnCredential{}, init...)}
}

type webauthnCredentialPersister struct {
	credentials []models.WebauthnCredential
}

func (p *webauthnCredentialPersister) Get(id string) (*models.WebauthnCredential, error) {
	var found *models.WebauthnCredential
	for _, data := range p.credentials {
		if data.ID == id {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (p *webauthnCredentialPersister) Create(credential models.WebauthnCredential) error {
	p.credentials = append(p.credentials, credential)
	return nil
}

func (p *webauthnCredentialPersister) Update(credential models.WebauthnCredential) error {
	for i, data := range p.credentials {
		if data.ID == credential.ID {
			p.credentials[i] = credential
		}
	}
	return nil
}

func (p *webauthnCredentialPersister) Delete(credential models.WebauthnCredential) error {
	index := -1
	for i, data := range p.credentials {
		if data.ID == credential.ID {
			index = i
		}
	}
	if index > -1 {
		p.credentials = append(p.credentials[:index], p.credentials[index+1:]...)
	}

	return nil
}

func (p *webauthnCredentialPersister) GetFromUser(id uuid.UUID) ([]models.WebauthnCredential, error) {
	var found []models.WebauthnCredential
	for _, data := range p.credentials {
		if data.UserId == id {
			found = append(found, data)
		}
	}
	return found, nil
}
