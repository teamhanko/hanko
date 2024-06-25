package flowpilot

import (
	"errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// JSONManagerPath represents a stash path.
type JSONManagerPath string

// ReadJSONManager is the interface that allows read operations.
type ReadJSONManager interface {
	Get(path JSONManagerPath) gjson.Result // Get retrieves the value at the specified path in the JSON data.
	String() string                        // String returns the JSON data as a string.
	Unmarshal() interface{}                // Unmarshal parses the JSON data and returns it as an interface{}.
}

// JSONManager is the interface that defines methods for reading, writing, and deleting JSON data.
type JSONManager interface {
	ReadJSONManager
	Set(path JSONManagerPath, value interface{}) error // Set updates the JSON data at the specified path with the provided value.
	Delete(path JSONManagerPath) error                 // Delete removes a value from the JSON data at the specified path.
}

// ReadOnlyJSONManager is the interface that allows only read operations.
type ReadOnlyJSONManager interface {
	ReadJSONManager
}

// DefaultJSONManager is the default implementation of the JSONManager interface.
type DefaultJSONManager struct {
	data string // The JSON data stored as a string.
}

// NewJSONManager creates a new instance of DefaultJSONManager with empty JSON data.
func NewJSONManager() JSONManager {
	return &DefaultJSONManager{data: "{}"}
}

// NewJSONManagerFromString creates a new instance of DefaultJSONManager with the given JSON data.
// It checks if the provided data is valid JSON before creating the instance.
func NewJSONManagerFromString(data string) (JSONManager, error) {
	if !gjson.Valid(data) {
		return nil, errors.New("invalid json")
	}
	return &DefaultJSONManager{data: data}, nil
}

// Get retrieves the value at the specified path in the JSON data.
func (jm *DefaultJSONManager) Get(path JSONManagerPath) gjson.Result {
	return gjson.Get(jm.data, string(path))
}

// Set updates the JSON data at the specified path with the provided value.
func (jm *DefaultJSONManager) Set(path JSONManagerPath, value interface{}) error {
	newData, err := sjson.Set(jm.data, string(path), value)
	if err != nil {
		return err
	}
	jm.data = newData
	return nil
}

// Delete removes a value from the JSON data at the specified path.
func (jm *DefaultJSONManager) Delete(path JSONManagerPath) error {
	newData, err := sjson.Delete(jm.data, string(path))
	if err != nil {
		return err
	}
	jm.data = newData
	return nil
}

// String returns the JSON data as a string.
func (jm *DefaultJSONManager) String() string {
	return jm.data
}

// Unmarshal parses the JSON data and returns it as an interface{}.
func (jm *DefaultJSONManager) Unmarshal() interface{} {
	m, ok := gjson.Parse(jm.data).Value().(interface{})
	if !ok {
		return nil
	}
	return m
}
