package flowpilot

import (
	"encoding/json"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

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

// newActionInputFromInputData creates a new instance of actionInput with the given JSON data
// which was previously unmarshalled into a generic map.
func newActionInputFromInputData(data InputData) (actionInput, error) {
	dataBytes, err := json.Marshal(data.InputDataMap)
	if err != nil {
		return nil, err
	}
	return jsonmanager.NewJSONManagerFromString(string(dataBytes))
}
