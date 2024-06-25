package flowpilot

import (
	"github.com/tidwall/gjson"
)

// initializationInputSchema represents an interface for managing input data schemas.
type initializationInputSchema interface {
	AddInputs(inputList ...Input)
}

// executionInputSchema represents an interface for managing method execution schemas.
type executionInputSchema interface {
	Get(path string) gjson.Result
	Set(path string, value interface{}) error
	SetError(inputName string, inputError InputError)

	getInput(name string) Input
	getOutputData() readOnlyActionInput
	getDataToPersist() readOnlyActionInput
	validateInputData(stateName StateName, stash stash) bool
	forInitializationContext() initializationInputSchema
	toResponse(stateName StateName) ResponseInputs
}

// inputs represents a collection of Input instances.
type inputs []Input

func (il *inputs) exists(input Input) bool {
	for _, existingInput := range *il {
		if existingInput.getName() == input.getName() {
			return true
		}
	}
	return false
}

// ResponseInputs represents a collection of ResponseInput instances.
type ResponseInputs map[string]*ResponseInput

// defaultSchema implements the initializationInputSchema interface and holds a collection of input fields.
type defaultSchema struct {
	inputs
	inputData  readOnlyActionInput
	outputData actionInput
}

// newSchemaWithInputData creates a new executionInputSchema with input data.
func newSchemaWithInputData(inputData actionInput) executionInputSchema {
	outputData := newActionInput()

	return &defaultSchema{
		inputData:  inputData,
		outputData: outputData,
	}
}

// newSchemaWithInputData creates a new executionInputSchema with input data.
func newSchemaWithOutputData(outputData readOnlyActionInput) (executionInputSchema, error) {
	data, err := newActionInputFromString(outputData.String())
	if err != nil {
		return nil, err
	}

	return &defaultSchema{
		inputData:  data,
		outputData: data,
	}, nil
}

// newSchema creates a new executionInputSchema with no input data.
func newSchema() executionInputSchema {
	inputData := newActionInput()
	return newSchemaWithInputData(inputData)
}

// toInitializationSchema converts executionInputSchema to initializationInputSchema.
func (s *defaultSchema) forInitializationContext() initializationInputSchema {
	return s
}

// Get retrieves a value at the specified path in the input data.
func (s *defaultSchema) Get(path string) gjson.Result {
	return s.inputData.Get(JSONManagerPath(path))
}

// Set updates the JSON data at the specified path with the provided value.
func (s *defaultSchema) Set(path string, value interface{}) error {
	return s.outputData.Set(JSONManagerPath(path), value)
}

// AddInputs adds input fields to the defaultSchema and returns the updated inputSchema.
func (s *defaultSchema) AddInputs(inputList ...Input) {
	for _, input := range inputList {
		if !s.inputs.exists(input) {
			s.inputs = append(s.inputs, input)
		}
	}
}

// getInput retrieves an input field from the inputSchema based on its name.
func (s *defaultSchema) getInput(name string) Input {
	for _, input := range s.inputs {
		if input.getName() == name {
			return input
		}
	}

	return nil
}

// SetError sets an error for an input field in the inputSchema.
func (s *defaultSchema) SetError(inputName string, inputError InputError) {
	if input := s.getInput(inputName); input != nil {
		input.setError(inputError)
	}
}

// validateInputData validates the input data based on the input definitions in the inputSchema.
func (s *defaultSchema) validateInputData(stateName StateName, stash stash) bool {
	valid := true

	for _, input := range s.inputs {
		if !input.validate(stateName, s.inputData, stash) && valid {
			valid = false
		}
	}

	return valid
}

// getDataToPersist filters and returns data that should be persisted based on inputSchema definitions.
func (s *defaultSchema) getDataToPersist() readOnlyActionInput {
	toPersist := newActionInput()

	for _, input := range s.inputs {
		if v := s.inputData.Get(JSONManagerPath(input.getName())); v.Exists() && input.shouldPersist() {
			_ = toPersist.Set(JSONManagerPath(input.getName()), v.Value())
		}
	}

	return toPersist
}

// getOutputData returns the output data from the inputSchema.
func (s *defaultSchema) getOutputData() readOnlyActionInput {
	return s.outputData
}

// toPublicSchema converts defaultSchema to ResponseInputs for public exposure.
func (s *defaultSchema) toResponse(stateName StateName) ResponseInputs {
	var publicSchema = make(ResponseInputs)

	for _, input := range s.inputs {
		if !input.isIncludedOnState(stateName) {
			continue
		}

		outputValue := s.outputData.Get(JSONManagerPath(input.getName()))
		inputValue := s.inputData.Get(JSONManagerPath(input.getName()))

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
