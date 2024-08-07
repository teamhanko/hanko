package test

import (
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewUsernamePersister(init []models.Username) persistence.UsernamePersister {
	return &usernamePersister{append([]models.Username{}, init...)}
}

type usernamePersister struct {
	usernames []models.Username
}

func (u *usernamePersister) Create(username models.Username) error {
	u.usernames = append(u.usernames, username)
	return nil
}

func (u *usernamePersister) GetByName(name string) (*models.Username, error) {
	var found *models.Username
	for _, data := range u.usernames {
		if data.Username == name {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (u *usernamePersister) Update(username *models.Username) error {
	for i, data := range u.usernames {
		if data.ID == username.ID {
			u.usernames[i] = *username
		}
	}
	return nil
}

func (u *usernamePersister) Delete(username *models.Username) error {
	index := -1
	for i, data := range u.usernames {
		if data.ID == username.ID {
			index = i
		}
	}
	if index > -1 {
		u.usernames = append(u.usernames[:index], u.usernames[index+1:]...)
	}

	return nil
}
