package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewUserPersister(init []models.User) persistence.UserPersister {
	return &userPersister{append([]models.User{}, init...)}
}

type userPersister struct {
	users []models.User
}

func (p *userPersister) Get(id uuid.UUID) (*models.User, error) {
	var found *models.User
	for _, data := range p.users {
		if data.ID == id {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (p *userPersister) GetByEmail(email string) (*models.User, error) {
	var found *models.User
	for _, data := range p.users {
		if data.Email == email {
			d := data
			found = &d
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

func (p *userPersister) List(page int, perPage int) ([]models.User, error) {
	if len(p.users) == 0 {
		return p.users, nil
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	var result [][]models.User
	var j int
	for i := 0; i < len(p.users); i += perPage {
		j += perPage
		if j > len(p.users) {
			j = len(p.users)
		}
		result = append(result, p.users[i:j])
	}

	if page > len(result) {
		return []models.User{}, nil
	}
	return result[page-1], nil
}

func (p *userPersister) Count() (int, error) {
	return len(p.users), nil
}
