package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewEmailPersister(init []models.Email) persistence.EmailPersister {
	return &emailPersister{append([]models.Email{}, init...)}
}

type emailPersister struct {
	emails []models.Email
}

func (e *emailPersister) FindByUserId(userId uuid.UUID) ([]models.Email, error) {
	var emails []models.Email
	for _, email := range e.emails {
		if email.UserID == userId {
			emails = append(emails, email)
		}
	}
	return emails, nil
}

func (e *emailPersister) FindByAddress(address string) (*models.Email, error) {
	for _, email := range e.emails {
		if email.Address == address {
			return &email, nil
		}
	}
	return nil, nil
}

func (e *emailPersister) Create(email models.Email) error {
	e.emails = append(e.emails, email)
	return nil
}

func (e *emailPersister) Update(email models.Email) error {
	for i, data := range e.emails {
		if data.ID == email.ID {
			e.emails[i] = email
		}
	}
	return nil
}

func (e *emailPersister) Delete(email models.Email) error {
	index := -1
	for i, data := range e.emails {
		if data.ID == email.ID {
			index = i
		}
	}
	if index > -1 {
		e.emails = append(e.emails[:index], e.emails[index+1:]...)
	}

	return nil
}
