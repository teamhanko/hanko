package test

import (
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewIdentityPersister(init []models.Identity) persistence.IdentityPersister {
	if init == nil {
		return &identityPersister{[]models.Identity{}}
	}
	return &identityPersister{append([]models.Identity{}, init...)}
}

type identityPersister struct {
	identities []models.Identity
}

func (i identityPersister) Get(userProviderID string, providerName string) (*models.Identity, error) {
	for _, identity := range i.identities {
		if identity.ProviderID == userProviderID && identity.ProviderName == providerName {
			return &identity, nil
		}
	}
	return nil, nil
}

func (i identityPersister) Create(identity models.Identity) error {
	i.identities = append(i.identities, identity)
	return nil
}

func (i identityPersister) Update(identity models.Identity) error {
	for idx, data := range i.identities {
		if data.ID == identity.ID {
			i.identities[idx] = identity
		}
	}
	return nil
}

func (i identityPersister) Delete(identity models.Identity) error {
	index := -1
	for idx, data := range i.identities {
		if data.ID == identity.ID {
			index = idx
		}
	}
	if index > -1 {
		i.identities = append(i.identities[:index], i.identities[index+1:]...)
	}

	return nil
}
