package flowpilot

import (
	"errors"
	"fmt"
)

// defaultActionExecutionContext is the default implementation of the actionExecutionContext interface.
type defaultActionExecutionContext struct {
	actionName          ActionName      // Name of the action being executed.
	input               ExecutionSchema // JSONManager for accessing input data.
	flowError           FlowError
	executionResult     *executionResult // Result of the action execution.
	links               []Link           // TODO:
	isSuspended         bool
	skipWriteHistory    bool
	*defaultFlowContext // Embedding the defaultFlowContext for common context fields.
}

// saveNextState updates the flow's state and stores data to the database.
func (aec *defaultActionExecutionContext) saveNextState(executionResult executionResult) error {
	newVersion := aec.flowModel.Version + 1
	stashData := aec.stash.String()
	previousState := aec.flowModel.CurrentState

	// Prepare parameters for updating the flow in the database.
	flowUpdate := flowUpdateParam{
		flowID:        aec.flowModel.ID,
		nextState:     executionResult.nextStateName,
		previousState: previousState,
		stashData:     stashData,
		version:       newVersion,
		csrfToken:     aec.csrfToken,
		expiresAt:     aec.flowModel.ExpiresAt,
		createdAt:     aec.flowModel.CreatedAt,
	}

	// Update the flow model in the database.
	if _, err := aec.dbw.updateFlowWithParam(flowUpdate); err != nil {
		return fmt.Errorf("failed to store updated flow: %w", err)
	}

	aec.flowModel.CurrentState = executionResult.nextStateName
	aec.flowModel.PreviousState = &previousState

	// Get the data to persists from the executed action schema for recording.
	inputDataToPersist := aec.input.getDataToPersist().String()

	// Prepare parameters for creating a new transition in the database.
	transitionCreation := transitionCreationParam{
		flowID:     aec.flowModel.ID,
		actionName: aec.actionName,
		fromState:  previousState,
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
func (aec *defaultActionExecutionContext) continueFlow(nextStateName StateName) error {
	// Retrieve the current state from the flow.
	currentState, err := aec.flow.getState(aec.flowModel.CurrentState)
	if err != nil {
		return fmt.Errorf("invalid current state: %w", err)
	}

	nextStateAllowed := currentState.getFlow().stateExists(nextStateName)

	// Check if the specified nextStateName is valid.
	if !(nextStateAllowed || nextStateName == aec.flow.errorStateName) {
		return fmt.Errorf("progression to the specified state '%s' is not allowed", nextStateName)
	}

	// Add the current state to the execution history.
	if currentState.getName() != nextStateName && !aec.skipWriteHistory {
		err = aec.stash.addStateToHistory(currentState.getName(), nil, nil)
		if err != nil {
			return fmt.Errorf("failed to add the current state to the history: %w", err)
		}
	}

	// Close the execution context with the given next state.
	return aec.closeExecutionContext(nextStateName)
}

func (aec *defaultActionExecutionContext) closeExecutionContext(nextStateName StateName) error {
	if aec.executionResult != nil {
		return errors.New("execution context is closed already")
	}

	if err := aec.executeAfterStateHooks(); err != nil {
		return fmt.Errorf("error while executing after hook actions: %w", err)
	}

	if err := aec.executeAfterEachActionHooks(); err != nil {
		return fmt.Errorf("error while executing after each action hook actions: %w", err)
	}

	if err := aec.executeBeforeStateHooks(nextStateName); err != nil {
		return fmt.Errorf("error while executing before hook actions: %w", err)
	}

	actionResult := actionExecutionResult{
		actionName:  aec.actionName,
		schema:      aec.input,
		isSuspended: aec.isSuspended,
	}

	result := executionResult{
		flowError:             aec.flowError,
		actionExecutionResult: &actionResult,
		links:                 aec.links,
		nextStateName:         nextStateName,
	}

	aec.executionResult = &result

	// Save the next state and transition data.
	if err := aec.saveNextState(result); err != nil {
		return fmt.Errorf("failed to save the transition data: %w", err)
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeBeforeStateHooks(nextStateName StateName) error {
	nextState, err := aec.flow.getState(nextStateName)
	if err != nil {
		return err
	}

	for _, hook := range nextState.getBeforeStateHooks().reverse() {
		err = hook.Execute(aec)
		if err != nil {
			return fmt.Errorf("failed to execute hook action before state '%s': %w", nextState.getName(), err)
		}
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeAfterStateHooks() error {
	currentState, err := aec.flow.getState(aec.flowModel.CurrentState)
	if err != nil {
		return err
	}

	for _, hook := range currentState.getAfterStateHooks().reverse() {
		err = hook.Execute(aec)
		if err != nil {
			return fmt.Errorf("failed to execute hook action after state: %w", err)
		}
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeBeforeEachActionHooks() error {
	for _, hook := range aec.flow.beforeEachActionHooks {
		err := hook.Execute(aec)
		if err != nil {
			return fmt.Errorf("failed to execute hook before action '%s'", aec.actionName)
		}
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeAfterEachActionHooks() error {
	for _, hook := range aec.flow.afterEachActionHooks {
		err := hook.Execute(aec)
		if err != nil {
			return fmt.Errorf("failed to execute hook before action '%s'", aec.actionName)
		}
	}

	return nil
}

// Input returns the ExecutionSchema for accessing input data.
func (aec *defaultActionExecutionContext) Input() ExecutionSchema {
	return aec.input
}

// payload returns the JSONManager for accessing payload data.
func (aec *defaultActionExecutionContext) Payload() payload {
	return aec.payload
}

// CopyInputValuesToStash copies specified inputs to the stash.
func (aec *defaultActionExecutionContext) CopyInputValuesToStash(inputNames ...string) error {
	for _, inputName := range inputNames {
		// Copy input values to the stash.
		if result := aec.input.Get(inputName); result.Exists() {
			if err := aec.stash.Set(inputName, result.Value()); err != nil {
				return err
			}
		}

	}

	return nil
}

func (aec *defaultActionExecutionContext) SetFlowError(err FlowError) {
	aec.flowError = err
}

func (aec *defaultActionExecutionContext) GetFlowError() FlowError {
	return aec.flowError
}

// ValidateInputData validates the input data against the schema.
func (aec *defaultActionExecutionContext) ValidateInputData() bool {
	return aec.input.validateInputData(aec.flowModel.CurrentState, aec.stash)
}

// ContinueFlow continues the flow execution to the specified nextStateName.
func (aec *defaultActionExecutionContext) ContinueFlow(nextStateName StateName) error {
	return aec.continueFlow(nextStateName)
}

// ContinueFlowWithError continues the flow execution to the specified nextStateName with an error type.
func (aec *defaultActionExecutionContext) ContinueFlowWithError(nextStateName StateName, flowErr FlowError) error {
	aec.flowError = flowErr
	return aec.continueFlow(nextStateName)
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

	// If going to the previous state "crosses" subflow boundaries, update stashed flowPath accordingly, i.e. remove
	// last subflow fragment so that going back and restarting the subflow again does not repeatedly amass the same
	// subflow name in the flowPath.
	subFlowToGoBackTo := aec.flow.subFlows.getSubFlowFromStateName(*lastStateName)
	currentSubFlow := aec.flow.subFlows.getSubFlowFromStateName(aec.flowModel.CurrentState)
	// If subFlowToGoBackTo is nil then we probably want to go back to a root flow state.
	if subFlowToGoBackTo == nil && currentSubFlow != nil ||
		(subFlowToGoBackTo != nil && currentSubFlow != nil) && subFlowToGoBackTo.getName() != currentSubFlow.getName() {
		newPath := newFlowPathFromString(aec.stash.Get("_.flowPath").String())
		newPath.remove()
		_ = aec.stash.Set("_.flowPath", newPath.String())
	}

	// Close the execution context with the last state.
	return aec.closeExecutionContext(*lastStateName)
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
	if !currentState.getSubFlows().stateExists(entryStateName) {
		return fmt.Errorf("the specified entry state '%s' is not associated with a sub-flow of the current flow", entryStateName)
	}

	var scheduledStates []StateName

	// Append valid states to the list of scheduledStates.
	for index, nextStateName := range nextStateNames {
		stateExists := currentState.getFlow().stateExists(nextStateName)
		subFlowStateExists := currentState.getSubFlows().stateExists(nextStateName)

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

	if !aec.skipWriteHistory {
		// Add the current state to the execution history.
		err = aec.stash.addStateToHistory(currentState.getName(), nil, &numOfScheduledStates)
		if err != nil {
			return fmt.Errorf("failed to add state to history: %w", err)
		}
	}

	sf := currentState.getSubFlows().getSubFlowFromStateName(entryStateName)
	newPath := newFlowPathFromString(aec.stash.Get("_.flowPath").String())
	newPath.add(sf.getName())
	err = aec.stash.Set("_.flowPath", newPath.String())
	if err != nil {
		return fmt.Errorf("failed to stash new flowPath: %w", err)
	}

	// Close the execution context with the entry state of the sub-flow.
	return aec.closeExecutionContext(entryStateName)
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
	} else if !aec.skipWriteHistory {
		// Add the current state to the execution history.
		err = aec.stash.addStateToHistory(currentStateName, scheduledStateName, nil)
		if err != nil {
			return fmt.Errorf("failed to add state to history: %w", err)
		}
	}

	newPath := newFlowPathFromString(aec.stash.Get("_.flowPath").String())
	newPath.remove()

	if scheduledSubflow := aec.flow.subFlows.getSubFlowFromStateName(*scheduledStateName); scheduledSubflow != nil {
		newPath.add(scheduledSubflow.getName())
	}

	err = aec.stash.Set("_.flowPath", newPath.String())
	if err != nil {
		return fmt.Errorf("failed to stash new flowPath: %w", err)
	}

	// Close the execution context with the scheduled state.
	return aec.closeExecutionContext(*scheduledStateName)
}

func (aec *defaultActionExecutionContext) AddLink(links ...Link) {
	aec.links = append(aec.links, links...)
}

func (aec *defaultActionExecutionContext) Set(key string, value interface{}) {
	aec.flow.Set(key, value)
}

func (aec *defaultActionExecutionContext) SuspendAction() {
	aec.isSuspended = true
}

func (aec *defaultActionExecutionContext) DeleteStateHistory(skipWriteHistory bool) error {
	aec.skipWriteHistory = skipWriteHistory
	return aec.stash.deleteStateHistory()
}
