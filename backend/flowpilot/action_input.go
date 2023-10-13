package flowpilot

import "github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"

type ActionInput interface {
	jsonmanager.JSONManager
}

type ReadOnlyActionInput interface {
	jsonmanager.ReadOnlyJSONManager
}

// NewActionInput creates a new instance of ActionInput with empty JSON data.
func NewActionInput() ActionInput {
	return jsonmanager.NewJSONManager()
}

func NewActionInputFromString(data string) (ActionInput, error) {
	return jsonmanager.NewJSONManagerFromString(data)
}
