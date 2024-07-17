package flowpilot

import (
	"github.com/tidwall/gjson"
	"strings"
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
	validateInputData() bool
	forInitializationContext() initializationInputSchema
	toResponseInputs() ResponseInputs
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
	inputData  actionInput
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
	return s.inputData.Get(path)
}

// Set updates the JSON data at the specified path with the provided value.
func (s *defaultSchema) Set(path string, value interface{}) error {
	return s.outputData.Set(path, value)
}

// AddInputs adds input fields to the defaultSchema and returns the updated inputSchema.
func (s *defaultSchema) AddInputs(inputList ...Input) {
	for _, input := range inputList {
		if !s.inputs.exists(input) {
			s.inputs = append(s.inputs, input)
		} else {
			for i, existingInput := range s.inputs {
				if existingInput.getName() == input.getName() {
					input.setError(existingInput.getError())
					s.inputs[i] = input
				}
			}
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
func (s *defaultSchema) validateInputData() bool {
	for _, input := range s.inputs {
		name := input.getName()

		if input.shouldTrimSpace() {
			v := strings.TrimSpace(s.inputData.Get(name).String())
			_ = s.inputData.Set(name, v)
		}

		if input.shouldConvertToLowerCase() {
			v := strings.ToLower(s.inputData.Get(name).String())
			_ = s.inputData.Set(name, v)
		}
	}

	valid := true

	for _, input := range s.inputs {
		if !input.validate(s.inputData) && valid {
			valid = false
		}
	}

	return valid
}

// getOutputData returns the output data from the inputSchema.
func (s *defaultSchema) getOutputData() readOnlyActionInput {
	return s.outputData
}

// toResponseInputs converts defaultSchema to ResponseInputs for public exposure.
func (s *defaultSchema) toResponseInputs() ResponseInputs {
	var publicSchema = make(ResponseInputs)

	for _, input := range s.inputs {
		//outputValue := s.outputData.Get(input.getName())
		//inputValue := s.inputData.Get(input.getName())
		//
		//if outputValue.Exists() {
		//	input.setValue(outputValue.Value())
		//}
		//
		//if input.shouldPreserve() && inputValue.Exists() && !outputValue.Exists() {
		//	input.setValue(inputValue.Value())
		//}

		publicSchema[input.getName()] = input.toResponseInput()
	}

	return publicSchema
}
