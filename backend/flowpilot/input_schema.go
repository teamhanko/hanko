package flowpilot

import (
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
	"github.com/tidwall/gjson"
)

// InitializationSchema represents an interface for managing input data schemas.
type InitializationSchema interface {
	AddInputs(inputList ...Input)
}

// MethodExecutionSchema represents an interface for managing method execution schemas.
type MethodExecutionSchema interface {
	Get(path string) gjson.Result
	Set(path string, value interface{}) error
	SetError(inputName string, inputError InputError)

	getInput(name string) Input
	getOutputData() jsonmanager.ReadOnlyJSONManager
	getDataToPersist() jsonmanager.ReadOnlyJSONManager
	validateInputData(stateName StateName, stash jsonmanager.JSONManager) bool
	toInitializationSchema() InitializationSchema
	toPublicSchema(stateName StateName) PublicSchema
}

// inputs represents a collection of Input instances.
type inputs []Input

// PublicSchema represents a collection of PublicInput instances.
type PublicSchema []*PublicInput

// defaultSchema implements the InitializationSchema interface and holds a collection of input fields.
type defaultSchema struct {
	inputs
	inputData  jsonmanager.ReadOnlyJSONManager
	outputData jsonmanager.JSONManager
}

// newSchemaWithInputData creates a new MethodExecutionSchema with input data.
func newSchemaWithInputData(inputData jsonmanager.ReadOnlyJSONManager) MethodExecutionSchema {
	outputData := jsonmanager.NewJSONManager()
	return &defaultSchema{
		inputData:  inputData,
		outputData: outputData,
	}
}

// newSchemaWithInputData creates a new MethodExecutionSchema with input data.
func newSchemaWithOutputData(outputData jsonmanager.ReadOnlyJSONManager) (MethodExecutionSchema, error) {
	data, err := jsonmanager.NewJSONManagerFromString(outputData.String())
	if err != nil {
		return nil, err
	}

	return &defaultSchema{
		inputData:  data,
		outputData: data,
	}, nil
}

// newSchema creates a new MethodExecutionSchema with no input data.
func newSchema() MethodExecutionSchema {
	inputData := jsonmanager.NewJSONManager()
	return newSchemaWithInputData(inputData)
}

// toInitializationSchema converts MethodExecutionSchema to InitializationSchema.
func (s *defaultSchema) toInitializationSchema() InitializationSchema {
	return s
}

// Get retrieves a value at the specified path in the input data.
func (s *defaultSchema) Get(path string) gjson.Result {
	return s.inputData.Get(path)
}

// Set updates the JSON data at the specified path with the provided value.
func (s *defaultSchema) Set(path string, value interface{}) error {
	return s.outputData.Set(path, value)
}

// AddInputs adds input fields to the defaultSchema and returns the updated schema.
func (s *defaultSchema) AddInputs(inputList ...Input) {
	for _, i := range inputList {
		s.inputs = append(s.inputs, i)
	}
}

// getInput retrieves an input field from the schema based on its name.
func (s *defaultSchema) getInput(name string) Input {
	for _, i := range s.inputs {
		if i.getName() == name {
			return i
		}
	}

	return nil
}

// SetError sets an error for an input field in the schema.
func (s *defaultSchema) SetError(inputName string, inputError InputError) {
	if i := s.getInput(inputName); i != nil {
		i.setError(inputError)
	}
}

// validateInputData validates the input data based on the input definitions in the schema.
func (s *defaultSchema) validateInputData(stateName StateName, stash jsonmanager.JSONManager) bool {
	valid := true

	for _, i := range s.inputs {
		if !i.validate(stateName, s.inputData, stash) && valid {
			valid = false
		}
	}

	return valid
}

// getDataToPersist filters and returns data that should be persisted based on schema definitions.
func (s *defaultSchema) getDataToPersist() jsonmanager.ReadOnlyJSONManager {
	toPersist := jsonmanager.NewJSONManager()

	for _, i := range s.inputs {
		if v := s.inputData.Get(i.getName()); v.Exists() && i.shouldPersist() {
			_ = toPersist.Set(i.getName(), v.Value())
		}
	}

	return toPersist
}

// getOutputData returns the output data from the schema.
func (s *defaultSchema) getOutputData() jsonmanager.ReadOnlyJSONManager {
	return s.outputData
}

// toPublicSchema converts defaultSchema to PublicSchema for public exposure.
func (s *defaultSchema) toPublicSchema(stateName StateName) PublicSchema {
	var pi PublicSchema

	for _, i := range s.inputs {
		if !i.isIncludedOnState(stateName) {
			continue
		}

		outputValue := s.outputData.Get(i.getName())
		inputValue := s.inputData.Get(i.getName())

		if outputValue.Exists() {
			i.setValue(outputValue.Value())
		}

		if i.shouldPreserve() && inputValue.Exists() && !outputValue.Exists() {
			i.setValue(inputValue.Value())
		}

		pi = append(pi, i.toPublicInput())
	}

	return pi
}
