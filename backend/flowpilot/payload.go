package flowpilot

import "github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"

type payload interface {
	jsonmanager.JSONManager
}

// NewPayload creates a new instance of Payload with empty JSON data.
func NewPayload() payload {
	return jsonmanager.NewJSONManager()
}
