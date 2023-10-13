package flowpilot

import (
	"errors"
	"fmt"
)

// defaultActionExecutionContext is the default implementation of the actionExecutionContext interface.
type defaultActionExecutionContext struct {
	actionName      ActionName       // Name of the action being executed.
	input           ExecutionSchema  // JSONManager for accessing input data.
	executionResult *executionResult // Result of the action execution.

	defaultFlowContext // Embedding the defaultFlowContext for common context fields.
}

// saveNextState updates the flow's state and saves Transition data after action execution.
func (aec *defaultActionExecutionContext) saveNextState(executionResult executionResult) error {
	completed := executionResult.nextState == aec.flow.endState
	newVersion := aec.flowModel.Version + 1
	stashData := aec.stash.String()

	// Prepare parameters for updating the flow in the database.
	flowUpdate := flowUpdateParam{
		flowID:    aec.flowModel.ID,
		nextState: executionResult.nextState,
		stashData: stashData,
		version:   newVersion,
		completed: completed,
		expiresAt: aec.flowModel.ExpiresAt,
		createdAt: aec.flowModel.CreatedAt,
	}

	// Update the flow model in the database.
	if _, err := aec.dbw.UpdateFlowWithParam(flowUpdate); err != nil {
		return fmt.Errorf("failed to store updated flow: %w", err)
	}

	// Get the data to persists from the executed action schema for recording.
	inputDataToPersist := aec.input.getDataToPersist().String()

	// Prepare parameters for creating a new Transition in the database.
	transitionCreation := transitionCreationParam{
		flowID:     aec.flowModel.ID,
		actionName: aec.actionName,
		fromState:  aec.flowModel.CurrentState,
		toState:    executionResult.nextState,
		inputData:  inputDataToPersist,
		flowError:  executionResult.flowError,
	}

	// Create a new Transition in the database.
	if _, err := aec.dbw.CreateTransitionWithParam(transitionCreation); err != nil {
		return fmt.Errorf("failed to store a new transition: %w", err)
	}

	return nil
}

// continueFlow continues the flow execution to the specified nextState with an optional error type.
func (aec *defaultActionExecutionContext) continueFlow(nextState StateName, flowError FlowError) error {
	currentState := aec.flowModel.CurrentState

	detail, err := aec.flow.getStateDetail(currentState)
	if err != nil {
		return fmt.Errorf("invalid current state: %w", err)
	}

	stateExists := detail.flow.stateExists(nextState)
	subFlowEntryStateAllowed := detail.subFlows.isEntryStateAllowed(nextState)

	// Check if the specified nextState is valid.
	if !(stateExists ||
		subFlowEntryStateAllowed ||
		nextState == aec.flow.endState ||
		nextState == aec.flow.errorState) {
		return fmt.Errorf("progression to the specified state '%s' is not allowed", nextState)
	}

	if currentState != nextState {
		err = aec.stash.addStateToHistory(currentState, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to add the current state to the history: %w", err)
		}
	}

	return aec.closeExecutionContext(nextState, flowError)
}

func (aec *defaultActionExecutionContext) closeExecutionContext(nextState StateName, flowError FlowError) error {
	if aec.executionResult != nil {
		return errors.New("execution context is closed already")
	}

	// Prepare the result for continuing the flow.
	actionResult := actionExecutionResult{
		actionName: aec.actionName,
		schema:     aec.input,
	}

	result := executionResult{
		nextState:             nextState,
		flowError:             flowError,
		actionExecutionResult: &actionResult,
	}

	// Save the next state and transition data.
	if err := aec.saveNextState(result); err != nil {
		return fmt.Errorf("failed to save the transition data: %w", err)
	}

	aec.executionResult = &result

	return nil
}

// Input returns the ExecutionSchema for accessing input data.
func (aec *defaultActionExecutionContext) Input() ExecutionSchema {
	return aec.input
}

// Payload returns the JSONManager for accessing payload data.
func (aec *defaultActionExecutionContext) Payload() Payload {
	return aec.payload
}

// CopyInputValuesToStash copies specified inputs to the stash.
func (aec *defaultActionExecutionContext) CopyInputValuesToStash(inputNames ...string) error {
	for _, inputName := range inputNames {
		// Copy input values to the stash.
		if err := aec.stash.Set(inputName, aec.input.Get(inputName).Value()); err != nil {
			return err
		}
	}

	return nil
}

// ValidateInputData validates the input data against the schema.
func (aec *defaultActionExecutionContext) ValidateInputData() bool {
	return aec.input.validateInputData(aec.flowModel.CurrentState, aec.stash)
}

// ContinueFlow continues the flow execution to the specified nextState.
func (aec *defaultActionExecutionContext) ContinueFlow(nextState StateName) error {
	return aec.continueFlow(nextState, nil)
}

// ContinueFlowWithError continues the flow execution to the specified nextState with an error type.
func (aec *defaultActionExecutionContext) ContinueFlowWithError(nextState StateName, flowErr FlowError) error {
	return aec.continueFlow(nextState, flowErr)
}

// ContinueToPreviousState continues the flow back to the previous state.
func (aec *defaultActionExecutionContext) ContinueToPreviousState() error {
	nextState, unscheduledState, numOfScheduledStates, err := aec.stash.getLastStateFromHistory()
	if err != nil {
		return fmt.Errorf("failed get last state from history: %w", err)
	}

	err = aec.stash.removeLastStateFromHistory()
	if err != nil {
		return fmt.Errorf("failed remove last state from history: %w", err)
	}

	if nextState == nil {
		nextState = &aec.flow.initialState
	}

	if unscheduledState != nil {
		err = aec.stash.addScheduledStates(*nextState)
		if err != nil {
			return fmt.Errorf("failed add scheduled states: %w", err)
		}
	}

	if numOfScheduledStates != nil {
		for range make([]struct{}, *numOfScheduledStates) {
			_, err = aec.stash.removeLastScheduledState()
			if err != nil {
				return fmt.Errorf("failed remove last scheduled state: %w", err)
			}
		}
	}

	return aec.closeExecutionContext(*nextState, nil)
}

// StartSubFlow initiates the sub-flow associated with the specified StateName of the entry state (first parameter).
// After a sub-flow action calls EndSubFlow(), the flow progresses to a state within the current flow or another
// sub-flow's entry state, as specified in the list of nextStates (every StateName passed after the first parameter).
func (aec *defaultActionExecutionContext) StartSubFlow(entryState StateName, nextStates ...StateName) error {
	currentState := aec.flowModel.CurrentState

	detail, err := aec.flow.getStateDetail(currentState)
	if err != nil {
		return fmt.Errorf("invalid current state: %w", err)
	}

	// the specified entry state must be an entry state to a sub-flow of the current flow
	if entryStateAllowed := detail.subFlows.isEntryStateAllowed(entryState); !entryStateAllowed {
		return fmt.Errorf("the specified entry state '%s' is not associated with a sub-flow of the current flow", entryState)
	}

	var scheduledStates []StateName

	// validate the specified nextStates and append valid state to the list of scheduledStates
	for index, nextState := range nextStates {
		stateExists := detail.flow.stateExists(nextState)
		subFlowEntryStateAllowed := detail.subFlows.isEntryStateAllowed(nextState)

		// validate the current next state
		if index == len(nextStates)-1 {
			// the last state must be a member of the current flow or a sub-flow entry state
			if !stateExists && !subFlowEntryStateAllowed {
				return fmt.Errorf("the last next state '%s' specified is not a sub-flow entry state or another state associated with the current flow", nextState)
			}
		} else {
			// every other state must be a sub-flow entry state
			if !subFlowEntryStateAllowed {
				return fmt.Errorf("the specified next state '%s' is not a sub-flow entry state of the current flow", nextState)
			}
		}

		// append the current nextState to the list of scheduled states
		scheduledStates = append(scheduledStates, nextState)
	}

	err = aec.stash.addScheduledStates(scheduledStates...)
	if err != nil {
		return fmt.Errorf("failed to stash scheduled states: %w", err)
	}

	numOfScheduledStates := int64(len(scheduledStates))

	err = aec.stash.addStateToHistory(currentState, nil, &numOfScheduledStates)
	if err != nil {
		return fmt.Errorf("failed to add state to history: %w", err)
	}

	return aec.closeExecutionContext(entryState, nil)
}

// EndSubFlow ends the current sub-flow and progresses the flow to the previously defined nextStates (see StartSubFlow()).
func (aec *defaultActionExecutionContext) EndSubFlow() error {
	currentState := aec.flowModel.CurrentState

	nextState, err := aec.stash.removeLastScheduledState()
	if err != nil {
		return fmt.Errorf("failed to end sub-flow: %w", err)
	}

	if nextState == nil {
		nextState = &aec.flow.endState
	} else {
		err = aec.stash.addStateToHistory(currentState, nextState, nil)
		if err != nil {
			return fmt.Errorf("failed to add state to history: %w", err)
		}
	}

	return aec.closeExecutionContext(*nextState, nil)
}
