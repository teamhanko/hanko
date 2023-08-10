package flowpilot

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

// Schema represents an interface for managing input data schemas.
type Schema interface {
	AddInputs(inputList ...*DefaultInput) *DefaultSchema
}

// MethodExecutionSchema represents an interface for managing method execution schemas.
type MethodExecutionSchema interface {
	SetError(inputName string, errType *ErrorType)
	toResponseSchema() ResponseSchema
}

// ResponseSchema represents an interface for response schemas.
type ResponseSchema interface {
	GetInput(name string) *DefaultInput
	preserveInputData(inputData jsonmanager.ReadOnlyJSONManager)
	getTransitionData(inputData jsonmanager.ReadOnlyJSONManager) (jsonmanager.ReadOnlyJSONManager, error)
	applyFlash(flashData jsonmanager.ReadOnlyJSONManager)
	applyStash(stashData jsonmanager.ReadOnlyJSONManager)
	validateInputData(inputData jsonmanager.ReadOnlyJSONManager, stash jsonmanager.ReadOnlyJSONManager) bool
	toPublicInputs() PublicInputs
}

// Inputs represents a collection of DefaultInput instances.
type Inputs []*DefaultInput

// PublicInputs represents a collection of PublicInput instances.
type PublicInputs []*PublicInput

// DefaultSchema implements the Schema interface and holds a collection of input fields.
type DefaultSchema struct {
	Inputs
}

// toResponseSchema converts the DefaultSchema to a ResponseSchema.
func (s *DefaultSchema) toResponseSchema() ResponseSchema {
	return s
}

// AddInputs adds input fields to the DefaultSchema and returns the updated schema.
func (s *DefaultSchema) AddInputs(inputList ...*DefaultInput) *DefaultSchema {
	for _, i := range inputList {
		s.Inputs = append(s.Inputs, i)
	}

	return s
}

// GetInput retrieves an input field from the schema based on its name.
func (s *DefaultSchema) GetInput(name string) *DefaultInput {
	for _, i := range s.Inputs {
		if i.name == name {
			return i
		}
	}

	return nil
}

// SetError sets an error type for an input field in the schema.
func (s *DefaultSchema) SetError(inputName string, errType *ErrorType) {
	if i := s.GetInput(inputName); i != nil {
		i.errorType = errType
	}
}

// validateInputData validates the input data based on the input definitions in the schema.
func (s *DefaultSchema) validateInputData(inputData jsonmanager.ReadOnlyJSONManager, stashData jsonmanager.ReadOnlyJSONManager) bool {
	valid := true

	for _, i := range s.Inputs {
		var inputValue *string
		var stashValue *string

		if v := inputData.Get(i.name); v.Exists() {
			inputValue = &v.Str
		}

		if v := stashData.Get(i.name); v.Exists() {
			stashValue = &v.Str
		}

		if !i.validate(inputValue, stashValue) && valid {
			valid = false
		}
	}

	return valid
}

// preserveInputData preserves input data by setting values of inputs that should be preserved.
func (s *DefaultSchema) preserveInputData(inputData jsonmanager.ReadOnlyJSONManager) {
	for _, i := range s.Inputs {
		if v := inputData.Get(i.name); v.Exists() {
			if i.preserveValue {
				i.setValue(v.Str)
			}
		}
	}
}

// getTransitionData filters input data to persist based on schema definitions.
func (s *DefaultSchema) getTransitionData(inputData jsonmanager.ReadOnlyJSONManager) (jsonmanager.ReadOnlyJSONManager, error) {
	toPersist := jsonmanager.NewJSONManager()

	for _, i := range s.Inputs {
		if v := inputData.Get(i.name); v.Exists() && i.persistValue {
			if err := toPersist.Set(i.name, v.Value()); err != nil {
				return nil, fmt.Errorf("failed to copy data: %v", err)
			}
		}
	}

	return toPersist, nil
}

// applyFlash updates input values in the schema with corresponding values from flash data.
func (s *DefaultSchema) applyFlash(flashData jsonmanager.ReadOnlyJSONManager) {
	for _, i := range s.Inputs {
		v := flashData.Get(i.name)

		if v.Exists() {
			i.setValue(v.Value())
		}
	}
}

// applyStash updates input values in the schema with corresponding values from stash data.
func (s *DefaultSchema) applyStash(stashData jsonmanager.ReadOnlyJSONManager) {
	n := 0

	for _, i := range s.Inputs {
		if !i.conditionalIncludeFromStash || stashData.Get(i.name).Exists() {
			s.Inputs[n] = i
			n++
		}
	}

	s.Inputs = s.Inputs[:n]
}

// toPublicInputs converts DefaultSchema to PublicInputs for public exposure.
func (s *DefaultSchema) toPublicInputs() PublicInputs {
	var pi PublicInputs

	for _, i := range s.Inputs {
		pi = append(pi, i.toPublicInput())
	}

	return pi
}
