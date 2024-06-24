package flowpilot

import "github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"

type actionInput interface {
	jsonmanager.JSONManager
}

type readOnlyActionInput interface {
	jsonmanager.ReadOnlyJSONManager
}

// newActionInput creates a new instance of actionInput with empty JSON data.
func newActionInput() actionInput {
	return jsonmanager.NewJSONManager()
}

// newActionInputFromString creates a new instance of actionInput with the given JSON data.
func newActionInputFromString(data string) (actionInput, error) {
	return jsonmanager.NewJSONManagerFromString(data)
}
