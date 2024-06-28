package flowpilot

import (
	"errors"
	"fmt"
	"net/http"
)

// defaultActionExecutionContext is the default implementation of the actionExecutionContext interface.
type defaultActionExecutionContext struct {
	actionName       ActionName           // Name of the action being executed.
	input            executionInputSchema // JSONManager for accessing input data.
	flowError        FlowError
	executionResult  *executionResult // Result of the action execution.
	links            []Link           // TODO:
	isSuspended      bool
	skipWriteHistory bool
	rewound          bool

	*defaultFlowContext // Embedding the defaultFlowContext for common context fields.
}

// saveNextState updates the flow's state and stores data to the database.
func (aec *defaultActionExecutionContext) saveNextState(executionResult *executionResult) error {
	var err error
	var previousState StateName

	if aec.rewound {
		lastState, _, _, _ := aec.stash.getLastStateFromHistory()
		if lastState != nil {
			previousState = *lastState
		}
	} else {
		if aec.flowModel.CurrentState != executionResult.nextStateName {
			previousState = aec.flowModel.CurrentState
		} else {
			previousState = *aec.flowModel.PreviousState
		}
	}

	csrfToken, err := generateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate csrf token: %w", err)
	}

	newVersion := aec.flowModel.Version + 1

	// Prepare parameters for updating the flow in the database.
	flowUpdate := flowUpdateParam{
		flowID:        aec.flowModel.ID,
		nextState:     executionResult.nextStateName,
		previousState: previousState,
		stashData:     aec.stash.String(),
		version:       newVersion,
		csrfToken:     csrfToken,
		expiresAt:     aec.flowModel.ExpiresAt,
		createdAt:     aec.flowModel.CreatedAt,
	}

	// Update the flow model in the database.
	if _, err = aec.dbw.updateFlowWithParam(flowUpdate); err != nil {
		return fmt.Errorf("failed to store updated flow: %w", err)
	}

	aec.flowModel.CurrentState = executionResult.nextStateName
	aec.flowModel.PreviousState = &previousState
	aec.flowModel.Version = newVersion
	aec.flowModel.CSRFToken = csrfToken

	// Get the data to persists from the executed action inputSchema for recording.
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
	if _, err = aec.dbw.createTransitionWithParam(transitionCreation); err != nil {
		return fmt.Errorf("failed to store a new transition: %w", err)
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeAllHooks(nextStateName *StateName) (*StateName, error) {
	// Helper function to execute after flow hooks and handle state progression
	executeAfterFlowHooks := func() (*StateName, error) {
		// Execute after flow hooks
		if err := aec.executeAfterFlowHooks(nextStateName); err != nil {
			return nil, fmt.Errorf("error while executing after flow hook actions: %w", err)
		}

		// Check if next state is nil or empty and try to get the last scheduled state
		if nextStateName == nil || len(*nextStateName) == 0 {
			if nextState, _ := aec.stash.removeLastScheduledState(); nextState == nil {
				return nil, ErrorFlowDiscontinuity.Wrap(errors.New("can't progress the flow because no scheduled states were available"))
			} else {
				nextStateName = nextState
			}
		}

		return nextStateName, nil
	}

	var err error

	// Execute the helper function to handle after flow hooks and state progression
	if nextStateName, err = executeAfterFlowHooks(); err != nil {
		return nil, err
	}

	// Execute after state hooks
	if err = aec.executeAfterStateHooks(); err != nil {
		return nil, fmt.Errorf("error while executing after state hook actions: %w", err)
	}

	// Execute after each action hooks
	if err = aec.executeAfterEachActionHooks(); err != nil {
		return nil, fmt.Errorf("error while executing after each action hook actions: %w", err)
	}

	// Execute before state hooks for the next state
	if err = aec.executeBeforeStateHooks(*nextStateName); err != nil {
		return nil, fmt.Errorf("error while executing before state hook actions: %w", err)
	}

	// Return the next state name
	return nextStateName, nil
}

func (aec *defaultActionExecutionContext) closeExecutionContext(nextStateName *StateName) error {
	var err error

	if aec.executionResult != nil {
		return errors.New("execution context is closed already")
	}

	ns, err := aec.executeAllHooks(nextStateName)
	if err != nil {
		return fmt.Errorf("hook execution error: %w", err)
	}

	nextStateName = ns

	actionResult := &actionExecutionResult{
		actionName:  aec.actionName,
		inputSchema: aec.input,
		isSuspended: aec.isSuspended,
	}

	result := &executionResult{
		flowError:             aec.flowError,
		actionExecutionResult: actionResult,
		links:                 aec.links,
		nextStateName:         *nextStateName,
	}

	aec.executionResult = result

	// Save the next state and transition data.
	if err = aec.saveNextState(result); err != nil {
		return fmt.Errorf("failed to save the transition data: %w", err)
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeBeforeStateHooks(nextStateName StateName) error {
	if actions := aec.flow.beforeStateHooks[nextStateName]; actions != nil {
		for _, hook := range actions.reverse() {
			if err := hook.Execute(aec); err != nil {
				return fmt.Errorf("failed to execute hook action before state '%s': %w", nextStateName, err)
			}
		}
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeAfterStateHooks() error {
	if actions := aec.flow.afterStateHooks[aec.flowModel.CurrentState]; actions != nil {
		for _, hook := range actions.reverse() {
			if err := hook.Execute(aec); err != nil {
				return fmt.Errorf("failed to execute hook action after flow '%s': %w", aec.flowModel.CurrentState, err)
			}
		}
	}

	return nil
}

func (aec *defaultActionExecutionContext) executeAfterFlowHooks(nextState *StateName) error {
	sdc, _ := aec.flow.getState(aec.flowModel.CurrentState)
	currentFlowName := sdc.getFlowName()

	shouldExecute := nextState == nil || len(*nextState) == 0
	if !shouldExecute && nextState != nil && len(*nextState) > 0 {
		sdn, _ := aec.flow.getState(*nextState)
		nextFlowName := sdn.getFlowName()
		shouldExecute = currentFlowName != nextFlowName
	}

	if !shouldExecute {
		return nil
	}

	actions := aec.flow.afterFlowHooks[currentFlowName]
	if actions == nil {
		return nil
	}

	for _, hook := range actions.reverse() {
		if err := hook.Execute(aec); err != nil {
			return fmt.Errorf("failed to execute hook after flow '%s': %w", currentFlowName, err)
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

// Input returns the executionInputSchema for accessing input data.
func (aec *defaultActionExecutionContext) Input() executionInputSchema {
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

// ValidateInputData validates the input data against the inputSchema.
func (aec *defaultActionExecutionContext) ValidateInputData() bool {
	return aec.input.validateInputData(aec.flowModel.CurrentState, aec.stash)
}

// Error continues the flow execution to the current state, if it's a bad request error or to the error state otherwise.
// The flow response will contain the given error.
func (aec *defaultActionExecutionContext) Error(err FlowError) error {
	aec.flowError = err

	if err.Status() == http.StatusBadRequest {
		return aec.Continue(aec.flowModel.CurrentState)
	}

	return aec.Continue(aec.flow.errorStateName)
}

// Back continues the flow back to the previous state.
func (aec *defaultActionExecutionContext) Back() error {
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

	aec.rewound = true

	// Close the execution context with the last state.
	return aec.closeExecutionContext(lastStateName)
}

func (aec *defaultActionExecutionContext) Continue(stateNames ...StateName) error {
	var nextState *StateName
	currentState := aec.flowModel.CurrentState

	for _, stateName := range stateNames {
		if _, ok := aec.flow.stateDetails[stateName]; !ok {
			return fmt.Errorf("cannot continue to state: %s", stateName)
		}
	}

	if len(stateNames) == 1 {
		nextState = &stateNames[0]

		// Add the current state to the execution history.
		if currentState != *nextState && !aec.skipWriteHistory {
			if err := aec.stash.addStateToHistory(currentState, nil, nil); err != nil {
				return fmt.Errorf("failed to add the current state to the history: %w", err)
			}
		}
	} else if len(stateNames) > 1 {
		nextState = &stateNames[0]
		scheduledStates := stateNames[1:]

		// Add the scheduled states to the stash.
		if err := aec.stash.addScheduledStates(scheduledStates...); err != nil {
			return fmt.Errorf("failed to stash scheduled states: %w", err)
		}

		if !aec.skipWriteHistory {
			statesToBeScheduled := stateNames[1:]
			numOfScheduledStated := int64(len(statesToBeScheduled))

			// Add the current state to the execution history.
			err := aec.stash.addStateToHistory(currentState, nil, &numOfScheduledStated)
			if err != nil {
				return fmt.Errorf("failed to add state to history: %w", err)
			}
		}
	} else {
		// Attempt to remove the last scheduled state from the stash.
		scheduledState, err := aec.stash.removeLastScheduledState()
		if err != nil {
			return fmt.Errorf("failed to end sub-flow: %w", err)
		}

		if !aec.skipWriteHistory && scheduledState != nil {
			// Add the current state to the execution history.
			err = aec.stash.addStateToHistory(currentState, scheduledState, nil)
			if err != nil {
				return fmt.Errorf("failed to add state to history: %w", err)
			}
		}

		nextState = scheduledState
	}

	return aec.closeExecutionContext(nextState)
}

func (aec *defaultActionExecutionContext) AddLink(links ...Link) {
	aec.links = append(aec.links, links...)
}

func (aec *defaultActionExecutionContext) ScheduleStates(stateNames ...StateName) error {
	// Add the scheduled states to the stash.
	if err := aec.stash.addScheduledStates(stateNames...); err != nil {
		return fmt.Errorf("failed to stash scheduled states: %w", err)
	}

	return nil
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
