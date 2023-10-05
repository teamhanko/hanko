package utils

import (
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

type Stash interface {
	jsonmanager.JSONManager
}

// NewStash creates a new instance of Stash with empty JSON data.
func NewStash() Stash {
	return jsonmanager.NewJSONManager()
}

// NewStashFromString creates a new instance of Stash with the given JSON data.
func NewStashFromString(data string) (Stash, error) {
	return jsonmanager.NewJSONManagerFromString(data)
}

type Payload interface {
	jsonmanager.JSONManager
}

// NewPayload creates a new instance of Payload with empty JSON data.
func NewPayload() Payload {
	return jsonmanager.NewJSONManager()
}

// NewPayloadFromString creates a new instance of Payload with the given JSON data.
func NewPayloadFromString(data string) (Payload, error) {
	return jsonmanager.NewJSONManagerFromString(data)
}

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
