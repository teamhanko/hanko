package flowpilot

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

// defaultMethodExecutionContext is the default implementation of the methodExecutionContext interface.
type defaultMethodExecutionContext struct {
	methodName   MethodName            // Name of the method being executed.
	input        MethodExecutionSchema // JSONManager for accessing input data.
	methodResult *executionResult      // Result of the method execution.

	defaultFlowContext // Embedding the defaultFlowContext for common context fields.
}

// saveNextState updates the Flow's state and saves Transition data after method execution.
func (mec *defaultMethodExecutionContext) saveNextState(executionResult executionResult) error {
	currentState := mec.flowModel.CurrentState
	previousState := mec.flowModel.PreviousState

	// Update the previous state only if the next state is different from the current state.
	if executionResult.nextState != currentState {
		previousState = &currentState
	}

	completed := executionResult.nextState == mec.flow.EndState
	newVersion := mec.flowModel.Version + 1

	// Prepare parameters for updating the Flow in the database.
	flowUpdateParam := flowUpdateParam{
		flowID:        mec.flowModel.ID,
		nextState:     executionResult.nextState,
		previousState: previousState,
		stashData:     mec.stash.String(),
		version:       newVersion,
		completed:     completed,
		expiresAt:     mec.flowModel.ExpiresAt,
		createdAt:     mec.flowModel.CreatedAt,
	}

	// Update the Flow model in the database.
	flowModel, err := mec.dbw.UpdateFlowWithParam(flowUpdateParam)
	if err != nil {
		return fmt.Errorf("failed to store updated flow: %w", err)
	}

	mec.flowModel = *flowModel

	// Get the data to persists from the executed method's schema for recording.
	inputDataToPersist := mec.input.getDataToPersist().String()

	// Prepare parameters for creating a new Transition in the database.
	transitionCreationParam := transitionCreationParam{
		flowID:     mec.flowModel.ID,
		methodName: mec.methodName,
		fromState:  currentState,
		toState:    executionResult.nextState,
		inputData:  inputDataToPersist,
		flowError:  executionResult.flowError,
	}

	// Create a new Transition in the database.
	_, err = mec.dbw.CreateTransitionWithParam(transitionCreationParam)
	if err != nil {
		return fmt.Errorf("failed to store a new transition: %w", err)
	}

	return nil
}

// continueFlow continues the Flow execution to the specified nextState with an optional error type.
func (mec *defaultMethodExecutionContext) continueFlow(nextState StateName, flowError FlowError) error {
	// Check if the specified nextState is valid.
	if exists := mec.flow.stateExists(nextState); !exists {
		return errors.New("the execution result contains an invalid state")
	}

	// Prepare an executionResult for continuing the Flow.
	methodResult := executionResult{
		nextState: nextState,
		flowError: flowError,
		methodExecutionResult: &methodExecutionResult{
			methodName: mec.methodName,
			schema:     mec.input,
		},
	}

	// Save the next state and transition data.
	err := mec.saveNextState(methodResult)
	if err != nil {
		return fmt.Errorf("failed to save the transition data: %w", err)
	}

	mec.methodResult = &methodResult

	return nil
}

// Input returns the MethodExecutionSchema for accessing input data.
func (mec *defaultMethodExecutionContext) Input() MethodExecutionSchema {
	return mec.input
}

// Payload returns the JSONManager for accessing payload data.
func (mec *defaultMethodExecutionContext) Payload() jsonmanager.JSONManager {
	return mec.payload
}

// CopyInputValuesToStash copies specified inputs to the stash.
func (mec *defaultMethodExecutionContext) CopyInputValuesToStash(inputNames ...string) error {
	for _, inputName := range inputNames {
		// Copy input values to the stash.
		err := mec.stash.Set(inputName, mec.input.Get(inputName).Value())
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateInputData validates the input data against the schema.
func (mec *defaultMethodExecutionContext) ValidateInputData() bool {
	return mec.input.validateInputData(mec.flowModel.CurrentState, mec.stash)
}

// ContinueFlow continues the Flow execution to the specified nextState.
func (mec *defaultMethodExecutionContext) ContinueFlow(nextState StateName) error {
	return mec.continueFlow(nextState, nil)
}

// ContinueFlowWithError continues the Flow execution to the specified nextState with an error type.
func (mec *defaultMethodExecutionContext) ContinueFlowWithError(nextState StateName, flowErr FlowError) error {
	return mec.continueFlow(nextState, flowErr)
}
