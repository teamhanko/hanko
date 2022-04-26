package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/persistence/models"
)

func NewUserPersister(init []models.User) persistence.UserPersister {
	return &userPersister{init}
}

type userPersister struct {
	users []models.User
}

func (p *userPersister) Get(id uuid.UUID) (*models.User, error) {
	var found *models.User
	for _, data := range p.users {
		if data.ID == id {
			found = &data
		}
	}
	return found, nil
}

func (p *userPersister) Create(user models.User) error {
	p.users = append(p.users, user)
	return nil
}

func (p *userPersister) Update(user models.User) error {
	for i, data := range p.users {
		if data.ID == user.ID {
			p.users[i] = user
		}
	}
	return nil
}

func (p *userPersister) Delete(user models.User) error {
	index := -1
	for i, data := range p.users {
		if data.ID == user.ID {
			index = i
		}
	}
	if index > -1 {
		p.users = append(p.users[:index], p.users[index+1:]...)
	}

	return nil
}
