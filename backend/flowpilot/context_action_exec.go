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
	links           []Link           // TODO:

	defaultFlowContext // Embedding the defaultFlowContext for common context fields.
}

// saveNextState updates the flow's state and stores data to the database.
func (aec *defaultActionExecutionContext) saveNextState(executionResult executionResult) error {
	newVersion := aec.flowModel.Version + 1
	stashData := aec.stash.String()

	// Prepare parameters for updating the flow in the database.
	flowUpdate := flowUpdateParam{
		flowID:    aec.flowModel.ID,
		nextState: executionResult.nextStateName,
		stashData: stashData,
		version:   newVersion,
		expiresAt: aec.flowModel.ExpiresAt,
		createdAt: aec.flowModel.CreatedAt,
	}

	// Update the flow model in the database.
	if _, err := aec.dbw.updateFlowWithParam(flowUpdate); err != nil {
		return fmt.Errorf("failed to store updated flow: %w", err)
	}

	// Get the data to persists from the executed action schema for recording.
	inputDataToPersist := aec.input.getDataToPersist().String()

	// Prepare parameters for creating a new transition in the database.
	transitionCreation := transitionCreationParam{
		flowID:     aec.flowModel.ID,
		actionName: aec.actionName,
		fromState:  aec.flowModel.CurrentState,
		toState:    executionResult.nextStateName,
		inputData:  inputDataToPersist,
		flowError:  executionResult.flowError,
	}

	// Create a new Transition in the database.
	if _, err := aec.dbw.createTransitionWithParam(transitionCreation); err != nil {
		return fmt.Errorf("failed to store a new transition: %w", err)
	}

	return nil
}

// continueFlow continues the flow execution to the specified nextStateName with an optional error type.
func (aec *defaultActionExecutionContext) continueFlow(nextStateName StateName, flowError FlowError) error {
	// Retrieve the current state from the flow.
	currentState, err := aec.flow.getState(aec.flowModel.CurrentState)
	if err != nil {
		return fmt.Errorf("invalid current state: %w", err)
	}

	nextStateAllowed := currentState.flow.stateExists(nextStateName)

	// Check if the specified nextStateName is valid.
	if !(nextStateAllowed || nextStateName == aec.flow.errorStateName) {
		return fmt.Errorf("progression to the specified state '%s' is not allowed", nextStateName)
	}

	// Add the current state to the execution history.
	if currentState.name != nextStateName {
		err = aec.stash.addStateToHistory(currentState.name, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to add the current state to the history: %w", err)
		}
	}

	// Close the execution context with the given next state.
	return aec.closeExecutionContext(nextStateName, flowError)
}

func (aec *defaultActionExecutionContext) closeExecutionContext(nextStateName StateName, flowError FlowError) error {
	if aec.executionResult != nil {
		return errors.New("execution context is closed already")
	}

	if err := aec.executeAfterHookActions(); err != nil {
		return fmt.Errorf("error while executing after hook actions: %w", err)
	}

	if err := aec.executeBeforeHookActions(nextStateName); err != nil {
		return fmt.Errorf("error while executing before hook actions: %w", err)
	}

	actionResult := actionExecutionResult{
		actionName: aec.actionName,
		schema:     aec.input,
	}

	result := executionResult{
		nextStateName:         nextStateName,
		flowError:             flowError,
		actionExecutionResult: &actionResult,
		links:                 aec.links,
	}

	aec.executionResult = &result

	// Save the next state and transition data.
	if err := aec.saveNextState(result); err != nil {
		return fmt.Errorf("failed to save the transition data: %w", err)
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeBeforeHookActions(nextStateName StateName) error {
	nextState, err := aec.flow.getState(nextStateName)
	if err != nil {
		return err
	}

	for _, hook := range nextState.beforeHooks {
		err = hook.Execute(aec)
		if err != nil {
			return fmt.Errorf("failed to execute hook action before state '%s': %w", nextState.name, err)
		}
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeAfterHookActions() error {
	currentState, err := aec.flow.getState(aec.flowModel.CurrentState)
	if err != nil {
		return err
	}

	for _, hook := range currentState.afterHooks {
		err = hook.Execute(aec)
		if err != nil {
			return fmt.Errorf("failed to execute hook action after state: %w", err)
		}
	}

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

func (aec *defaultActionExecutionContext) SetFlowError(err FlowError) {
	aec.executionResult.flowError = err
}

func (aec *defaultActionExecutionContext) GetFlowError() FlowError {
	return aec.executionResult.flowError
}

// ValidateInputData validates the input data against the schema.
func (aec *defaultActionExecutionContext) ValidateInputData() bool {
	return aec.input.validateInputData(aec.flowModel.CurrentState, aec.stash)
}

// ContinueFlow continues the flow execution to the specified nextStateName.
func (aec *defaultActionExecutionContext) ContinueFlow(nextStateName StateName) error {
	return aec.continueFlow(nextStateName, nil)
}

// ContinueFlowWithError continues the flow execution to the specified nextStateName with an error type.
func (aec *defaultActionExecutionContext) ContinueFlowWithError(nextStateName StateName, flowErr FlowError) error {
	return aec.continueFlow(nextStateName, flowErr)
}

// ContinueToPreviousState continues the flow back to the previous state.
func (aec *defaultActionExecutionContext) ContinueToPreviousState() error {
	// Get the last state, the unscheduled state, and the number of scheduled states from history.
	lastStateName, unscheduledState, numOfScheduledStates, err := aec.stash.getLastStateFromHistory()
	if err != nil {
		return fmt.Errorf("failed get last state from history: %w", err)
	}

	// Remove the last state from history.
	err = aec.stash.removeLastStateFromHistory()
	if err != nil {
		return fmt.Errorf("failed remove last state from history: %w", err)
	}

	// If there was no last state, set it to the initial state.
	if lastStateName == nil {
		lastStateName = &aec.flow.initialStateName
	}

	// Add the unscheduled state back to the scheduled states if available.
	if unscheduledState != nil {
		err = aec.stash.addScheduledStates(*unscheduledState)
		if err != nil {
			return fmt.Errorf("failed add scheduled states: %w", err)
		}
	}

	// Remove any previously scheduled states from the schedule.
	if numOfScheduledStates != nil {
		for range make([]struct{}, *numOfScheduledStates) {
			_, err = aec.stash.removeLastScheduledState()
			if err != nil {
				return fmt.Errorf("failed remove last scheduled state: %w", err)
			}
		}
	}

	// Close the execution context with the last state.
	return aec.closeExecutionContext(*lastStateName, nil)
}

// StartSubFlow initiates a sub-flow associated with the specified entryStateName (first parameter). When a sub-flow
// action calls EndSubFlow(), the flow progresses to a state within the current flow or another sub-flow's entry state,
// as specified in the list of nextStates (every StateName passed after the first parameter).
func (aec *defaultActionExecutionContext) StartSubFlow(entryStateName StateName, nextStateNames ...StateName) error {
	// Retrieve the current state from the flow.
	currentState, err := aec.flow.getState(aec.flowModel.CurrentState)
	if err != nil {
		return fmt.Errorf("invalid current state: %w", err)
	}

	// Ensure the specified entry state is associated with a sub-flow of the current flow.
	if !currentState.subFlows.stateExists(entryStateName) {
		return fmt.Errorf("the specified entry state '%s' is not associated with a sub-flow of the current flow", entryStateName)
	}

	var scheduledStates []StateName

	// Append valid states to the list of scheduledStates.
	for index, nextStateName := range nextStateNames {
		stateExists := currentState.flow.stateExists(nextStateName)
		subFlowStateExists := currentState.subFlows.stateExists(nextStateName)

		// Validate the current next state.
		if index == len(nextStateNames)-1 {
			// The last state must be a member of the current flow or a sub-flow.
			if !stateExists && !subFlowStateExists {
				return fmt.Errorf("the last next state '%s' specified is not a sub-flow state or another state associated with the current flow", nextStateName)
			}
		} else {
			// Every other state must be a sub-flow state.
			if !subFlowStateExists {
				return fmt.Errorf("the specified next state '%s' is not a sub-flow state of the current flow", nextStateName)
			}
		}

		// Append the current next state to the list of scheduled states.
		scheduledStates = append(scheduledStates, nextStateName)
	}

	// Add the scheduled states to the stash.
	err = aec.stash.addScheduledStates(scheduledStates...)
	if err != nil {
		return fmt.Errorf("failed to stash scheduled states: %w", err)
	}

	numOfScheduledStates := int64(len(scheduledStates))

	// Add the current state to the execution history.
	err = aec.stash.addStateToHistory(currentState.name, nil, &numOfScheduledStates)
	if err != nil {
		return fmt.Errorf("failed to add state to history: %w", err)
	}

	// Close the execution context with the entry state of the sub-flow.
	return aec.closeExecutionContext(entryStateName, nil)
}

// EndSubFlow ends the current sub-flow and progresses the flow to the previously defined next states.
func (aec *defaultActionExecutionContext) EndSubFlow() error {
	// Retrieve the name of the current state.
	currentStateName := aec.flowModel.CurrentState

	// Attempt to remove the last scheduled state from the stash.
	scheduledStateName, err := aec.stash.removeLastScheduledState()
	if err != nil {
		return fmt.Errorf("failed to end sub-flow: %w", err)
	}

	// If no scheduled state is available, set it to the end state.
	if scheduledStateName == nil {
		return ErrorFlowDiscontinuity.Wrap(errors.New("can't progress the flow, because no scheduled states were available after the sub-flow ended"))
	} else {
		// Add the current state to the execution history.
		err = aec.stash.addStateToHistory(currentStateName, scheduledStateName, nil)
		if err != nil {
			return fmt.Errorf("failed to add state to history: %w", err)
		}
	}

	// Close the execution context with the scheduled state.
	return aec.closeExecutionContext(*scheduledStateName, nil)
}

func (aec *defaultActionExecutionContext) AddLink(links ...Link) {
	aec.links = append(aec.links, links...)
}
