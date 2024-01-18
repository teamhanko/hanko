package flowpilot

import (
	"github.com/tidwall/gjson"
)

// InitializationSchema represents an interface for managing input data schemas.
type InitializationSchema interface {
	AddInputs(inputList ...Input)
}

// ExecutionSchema represents an interface for managing method execution schemas.
type ExecutionSchema interface {
	Get(path string) gjson.Result
	Set(path string, value interface{}) error
	SetError(inputName string, inputError InputError)

	getInput(name string) Input
	getOutputData() ReadOnlyActionInput
	getDataToPersist() ReadOnlyActionInput
	validateInputData(stateName StateName, stash Stash) bool
	toInitializationSchema() InitializationSchema
	toPublicSchema(stateName StateName) PublicSchema
}

// inputs represents a collection of Input instances.
type inputs []Input

// PublicSchema represents a collection of PublicInput instances.
type PublicSchema map[string]*PublicInput

// defaultSchema implements the InitializationSchema interface and holds a collection of input fields.
type defaultSchema struct {
	inputs
	inputData  ReadOnlyActionInput
	outputData ActionInput
}

// newSchemaWithInputData creates a new ExecutionSchema with input data.
func newSchemaWithInputData(inputData ActionInput) ExecutionSchema {
	outputData := NewActionInput()

	return &defaultSchema{
		inputData:  inputData,
		outputData: outputData,
	}
}

// newSchemaWithInputData creates a new ExecutionSchema with input data.
func newSchemaWithOutputData(outputData ReadOnlyActionInput) (ExecutionSchema, error) {
	data, err := NewActionInputFromString(outputData.String())
	if err != nil {
		return nil, err
	}

	return &defaultSchema{
		inputData:  data,
		outputData: data,
	}, nil
}

// newSchema creates a new ExecutionSchema with no input data.
func newSchema() ExecutionSchema {
	inputData := NewActionInput()
	return newSchemaWithInputData(inputData)
}

// toInitializationSchema converts ExecutionSchema to InitializationSchema.
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
	for _, input := range inputList {
		s.inputs = append(s.inputs, input)
	}
}

// getInput retrieves an input field from the schema based on its name.
func (s *defaultSchema) getInput(name string) Input {
	for _, input := range s.inputs {
		if input.getName() == name {
			return input
		}
	}

	return nil
}

// SetError sets an error for an input field in the schema.
func (s *defaultSchema) SetError(inputName string, inputError InputError) {
	if input := s.getInput(inputName); input != nil {
		input.setError(inputError)
	}
}

// validateInputData validates the input data based on the input definitions in the schema.
func (s *defaultSchema) validateInputData(stateName StateName, stash Stash) bool {
	valid := true

	for _, input := range s.inputs {
		if !input.validate(stateName, s.inputData, stash) && valid {
			valid = false
		}
	}

	return valid
}

// getDataToPersist filters and returns data that should be persisted based on schema definitions.
func (s *defaultSchema) getDataToPersist() ReadOnlyActionInput {
	toPersist := NewActionInput()

	for _, input := range s.inputs {
		if v := s.inputData.Get(input.getName()); v.Exists() && input.shouldPersist() {
			_ = toPersist.Set(input.getName(), v.Value())
		}
	}

	return toPersist
}

// getOutputData returns the output data from the schema.
func (s *defaultSchema) getOutputData() ReadOnlyActionInput {
	return s.outputData
}

// toPublicSchema converts defaultSchema to PublicSchema for public exposure.
func (s *defaultSchema) toPublicSchema(stateName StateName) PublicSchema {
	var publicSchema = make(PublicSchema)

	for _, input := range s.inputs {
		if !input.isIncludedOnState(stateName) {
			continue
		}

		outputValue := s.outputData.Get(input.getName())
		inputValue := s.inputData.Get(input.getName())

		if outputValue.Exists() {
			input.setValue(outputValue.Value())
		}

		if input.shouldPreserve() && inputValue.Exists() && !outputValue.Exists() {
			input.setValue(inputValue.Value())
		}

		publicSchema[input.getName()] = input.toPublicInput()
	}

	return publicSchema
}
