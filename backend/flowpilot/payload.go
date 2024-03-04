package flowpilot

import "github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"

type Payload interface {
	jsonmanager.JSONManager
}

// NewPayload creates a new instance of Payload with empty JSON data.
func NewPayload() Payload {
	return jsonmanager.NewJSONManager()
}
