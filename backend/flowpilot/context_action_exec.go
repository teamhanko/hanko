package flowpilot

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
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
func (aec *defaultActionExecutionContext) continueFlow(nextState StateName, flowError FlowError, skipFlowValidation bool) error {
	detail, err := aec.flow.getStateDetail(aec.flowModel.CurrentState)
	if err != nil {
		return err
	}

	if !skipFlowValidation {
		// Check if the specified nextState is valid.
		if _, ok := detail.flow[nextState]; !(ok ||
			detail.subFlows.isEntryStateAllowed(nextState) ||
			nextState == aec.flow.endState ||
			nextState == aec.flow.errorState) {
			return fmt.Errorf("progression to the specified state '%s' is not allowed", nextState)
		}
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
func (aec *defaultActionExecutionContext) Payload() utils.Payload {
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
	return aec.continueFlow(nextState, nil, false)
}

// ContinueFlowWithError continues the flow execution to the specified nextState with an error type.
func (aec *defaultActionExecutionContext) ContinueFlowWithError(nextState StateName, flowErr FlowError) error {
	return aec.continueFlow(nextState, flowErr, false)
}

// StartSubFlow initiates the sub-flow associated with the specified StateName of the entry state (first parameter).
// After a sub-flow action calls EndSubFlow(), the flow progresses to a state within the current flow or another
// sub-flow's entry state, as specified in the list of nextStates (every StateName passed after the first parameter).
func (aec *defaultActionExecutionContext) StartSubFlow(entryState StateName, nextStates ...StateName) error {
	detail, err := aec.flow.getStateDetail(aec.flowModel.CurrentState)
	if err != nil {
		return err
	}

	// the specified entry state must be an entry state to a sub-flow of the current flow
	if entryStateAllowed := detail.subFlows.isEntryStateAllowed(entryState); !entryStateAllowed {
		return errors.New("the specified entry state is not associated with a sub-flow of the current flow")
	}

	var scheduledStates []StateName

	for index, nextState := range nextStates {
		subFlowEntryStateAllowed := detail.subFlows.isEntryStateAllowed(nextState)

		// validate the current next state
		if index == len(nextStates)-1 {
			// the last state must be a member of the current flow or a sub-flow entry state
			if _, ok := detail.flow[nextState]; !ok && !subFlowEntryStateAllowed {
				return errors.New("the last next state is not a sub-flow entry state or a state associated with the current flow")
			}
		} else {
			// every other state must be a sub-flow entry state
			if !subFlowEntryStateAllowed {
				return fmt.Errorf("next state with index %d is not a sub-flow entry state", index)
			}
		}

		// append the current nextState to the list of scheduled states
		scheduledStates = append(scheduledStates, nextState)
	}

	// get the current sub-flow stack from the stash
	stack := aec.stash.Get("_.scheduled_states").Array()

	newStack := make([]StateName, len(stack))

	for index := range newStack {
		newStack[index] = StateName(stack[index].String())
	}

	// prepend the states to the list of previously defined scheduled states
	newStack = append(scheduledStates, newStack...)

	err = aec.stash.Set("_.scheduled_states", newStack)
	if err != nil {
		return fmt.Errorf("failed to stash scheduled states while staring a sub-flow: %w", err)
	}

	return aec.continueFlow(entryState, nil, false)
}

// EndSubFlow ends the current sub-flow and progresses the flow to the previously defined nextStates (see StartSubFlow)
func (aec *defaultActionExecutionContext) EndSubFlow() error {
	// retrieve the previously scheduled states form the stash
	stack := aec.stash.Get("_.scheduled_states").Array()

	newStack := make([]StateName, len(stack))

	for index := range newStack {
		newStack[index] = StateName(stack[index].String())
	}

	// if there is no scheduled state left, continue to the end state
	if len(newStack) == 0 {
		newStack = append(newStack, aec.GetEndState())
	}

	// get and remove first stack item
	nextState := newStack[0]
	newStack = newStack[1:]

	// stash the updated list of scheduled states
	if err := aec.stash.Set("_.scheduled_states", newStack); err != nil {
		return fmt.Errorf("failed to stash scheduled states while ending the sub-flow: %w", err)
	}

	return aec.continueFlow(nextState, nil, true)
}
