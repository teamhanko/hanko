package test

import (
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewJwkPersister(init []models.Jwk) persistence.JwkPersister {
	if init == nil {
		return &jwkPersister{[]models.Jwk{}}
	}
	return &jwkPersister{append([]models.Jwk{}, init...)}
}

type jwkPersister struct {
	keys []models.Jwk
}

func (j *jwkPersister) Get(id int) (*models.Jwk, error) {
	var found *models.Jwk
	for _, data := range j.keys {
		if data.ID == id {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (j *jwkPersister) GetAll() ([]models.Jwk, error) {
	return j.keys, nil
}

func (j *jwkPersister) GetLast() (*models.Jwk, error) {
	l := len(j.keys)
	if l == 0 {
		return nil, nil
	}
	return &j.keys[l-1], nil
}

func (j *jwkPersister) Create(jwk models.Jwk) error {
	lastId := 0
	for _, key := range j.keys {
		if key.ID > lastId {
			lastId = key.ID
		}
	}
	jwk.ID = lastId
	j.keys = append(j.keys, jwk)
	return nil
}
