package flowpilot

import (
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

// Stash defines the interface for managing JSON data.
type Stash interface {
	stateHistoryStash
	scheduledStatesStash
	jsonmanager.JSONManager
}

// defaultStash implements the Stash interface.
type defaultStash struct {
	stateHistoryStash
	scheduledStatesStash
	jsonmanager.JSONManager
}

// newStashFromJSONManager creates a new instance of Stash with a given JSONManager.
func newStashFromJSONManager(jm jsonmanager.JSONManager) Stash {
	return &defaultStash{
		JSONManager:          jm,
		stateHistoryStash:    &defaultStateHistoryStash{JSONManager: jm},
		scheduledStatesStash: &defaultScheduledStatesStash{JSONManager: jm},
	}
}

// NewStash creates a new instance of Stash with empty JSON data.
func NewStash() Stash {
	jm := jsonmanager.NewJSONManager()
	return newStashFromJSONManager(jm)
}

// NewStashFromString creates a new instance of Stash with the given JSON data.
func NewStashFromString(data string) (Stash, error) {
	jm, err := jsonmanager.NewJSONManagerFromString(data)
	return newStashFromJSONManager(jm), err
}
