package test

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewSamlStatePersister(init []models.SamlState) persistence.SamlStatePersister {
	return &samlStatePersister{append([]models.SamlState{}, init...)}
}

type samlStatePersister struct {
	samlStates []models.SamlState
}

func (s samlStatePersister) Create(state models.SamlState) error {
	s.samlStates = append(s.samlStates, state)

	return nil
}

func (s samlStatePersister) GetByNonce(nonce string) (*models.SamlState, error) {
	for _, state := range s.samlStates {
		if state.Nonce == nonce {
			return &state, nil
		}
	}

	return nil, fmt.Errorf("failed to get state by nonce: %s", nonce)
}

func (s samlStatePersister) Delete(state models.SamlState) error {
	index := -1
	for i, existingState := range s.samlStates {
		if existingState.ID == state.ID {
			index = i
		}
	}
	if index > -1 {
		s.samlStates = append(s.samlStates[:index], s.samlStates[index+1:]...)
	}

	return nil
}
