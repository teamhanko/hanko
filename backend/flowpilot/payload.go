package flowpilot

type payload interface {
	JSONManager
}

// newPayload creates a new instance of Payload with empty JSON data.
func newPayload() payload {
	return NewJSONManager()
}
