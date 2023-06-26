package test

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

func NewOidcAuthRequestsPersister(init []models.AuthRequest, codes map[string]uuid.UUID) persistence.OIDCAuthRequestPersister {
	return &oidcAuthRequestsPersister{
		oidcAuthRequests: append([]models.AuthRequest{}, init...),
		oidcAuthCodes:    codes,
	}
}

type oidcAuthRequestsPersister struct {
	oidcAuthRequests []models.AuthRequest
	oidcAuthCodes    map[string]uuid.UUID
}

func (o *oidcAuthRequestsPersister) Get(ctx context.Context, uuid uuid.UUID) (*models.AuthRequest, error) {
	var found *models.AuthRequest

	for _, data := range o.oidcAuthRequests {
		if data.ID == uuid {
			d := data
			found = &d
		}
	}

	return found, nil
}

func (o *oidcAuthRequestsPersister) Create(ctx context.Context, authRequest models.AuthRequest) error {
	o.oidcAuthRequests = append(o.oidcAuthRequests, authRequest)

	return nil
}

func (o *oidcAuthRequestsPersister) Delete(ctx context.Context, uuid uuid.UUID) error {
	index := -1

	for i, data := range o.oidcAuthRequests {
		if data.ID == uuid {
			index = i
		}
	}

	if index > -1 {
		o.oidcAuthRequests = append(o.oidcAuthRequests[:index], o.oidcAuthRequests[index+1:]...)
	}

	for code, id := range o.oidcAuthCodes {
		if id == uuid {
			delete(o.oidcAuthCodes, code)
		}
	}

	return nil
}

func (o *oidcAuthRequestsPersister) AuthorizeUser(ctx context.Context, uuid uuid.UUID, userID string) error {
	for i, data := range o.oidcAuthRequests {
		if data.ID == uuid {
			o.oidcAuthRequests[i].UserID = userID
			o.oidcAuthRequests[i].Done = true
			o.oidcAuthRequests[i].AuthTime = time.Now()
		}
	}

	return nil
}

func (o *oidcAuthRequestsPersister) StoreAuthCode(ctx context.Context, ID uuid.UUID, code string) error {
	o.oidcAuthCodes[code] = ID

	return nil
}

func (o *oidcAuthRequestsPersister) GetAuthRequestByCode(ctx context.Context, code string) (*models.AuthRequest, error) {
	var found *models.AuthRequest

	uid, ok := o.oidcAuthCodes[code]
	if !ok {
		return nil, nil
	}

	for _, data := range o.oidcAuthRequests {
		if data.ID == uid {
			d := data
			found = &d
		}
	}

	return found, nil
}
