package flowpilot

type actionInput interface {
	JSONManager
}

type readOnlyActionInput interface {
	ReadOnlyJSONManager
}

// newActionInput creates a new instance of actionInput with empty JSON data.
func newActionInput() actionInput {
	return NewJSONManager()
}

// newActionInputFromString creates a new instance of actionInput with the given JSON data.
func newActionInputFromString(data string) (actionInput, error) {
	return NewJSONManagerFromString(data)
}
