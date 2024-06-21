package flowpilot

import (
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

// Stash defines the interface for managing JSON data.
type stash interface {
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

// newStashFromJSONManager creates a new instance of stash with a given JSONManager.
func newStashFromJSONManager(jm jsonmanager.JSONManager) stash {
	return &defaultStash{
		JSONManager:          jm,
		stateHistoryStash:    &defaultStateHistoryStash{JSONManager: jm},
		scheduledStatesStash: &defaultScheduledStatesStash{JSONManager: jm},
	}
}

// newStash creates a new instance of Stash with empty JSON data.
func newStash() stash {
	jm := jsonmanager.NewJSONManager()
	return newStashFromJSONManager(jm)
}

// newStashFromString creates a new instance of Stash with the given JSON data.
func newStashFromString(data string) (stash, error) {
	jm, err := jsonmanager.NewJSONManagerFromString(data)
	return newStashFromJSONManager(jm), err
}
