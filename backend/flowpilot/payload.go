package flowpilot

import "github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"

type payload interface {
	jsonmanager.JSONManager
}

// newPayload creates a new instance of Payload with empty JSON data.
func newPayload() payload {
	return jsonmanager.NewJSONManager()
}
