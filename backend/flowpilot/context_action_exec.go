package flowpilot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// defaultActionExecutionContext is the default implementation of the actionExecutionContext interface.
type defaultActionExecutionContext struct {
	actionName      ActionName           // Name of the action being executed.
	inputSchema     executionInputSchema // JSONManager for accessing input data.
	flowError       FlowError
	executionResult *executionResult // Result of the action execution.
	links           []Link           // TODO:
	isSuspended     bool
	preventRevert   bool

	*defaultFlowContext // Embedding the defaultFlowContext for common context fields.
}

// closeExecutionContext updates the flow's state and stores data to the database.
func (aec *defaultActionExecutionContext) closeExecutionContext() error {
	var err error

	if aec.executionResult != nil {
		return errors.New("execution context is closed already")
	}

	nextStateName := aec.stash.getStateName()

	actionResult := &actionExecutionResult{
		actionName:  aec.actionName,
		inputSchema: aec.inputSchema,
		isSuspended: aec.isSuspended,
	}

	result := &executionResult{
		flowError:             aec.flowError,
		actionExecutionResult: actionResult,
		links:                 aec.links,
		nextStateName:         nextStateName,
	}

	aec.executionResult = result

	csrfToken, err := generateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate csrf token: %w", err)
	}

	newVersion := aec.flowModel.Version + 1

	// Prepare parameters for updating the flow in the database.
	flowUpdate := flowUpdateParam{
		flowID:    aec.flowModel.ID,
		data:      aec.stash.String(),
		version:   newVersion,
		csrfToken: csrfToken,
		expiresAt: aec.flowModel.ExpiresAt,
		createdAt: aec.flowModel.CreatedAt,
	}

	// Update the flow model in the database.
	if _, err = aec.dbw.updateFlowWithParam(flowUpdate); err != nil {
		return fmt.Errorf("failed to store updated flow: %w", err)
	}

	aec.flowModel.Version = newVersion
	aec.flowModel.CSRFToken = csrfToken

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

func (aec *defaultActionExecutionContext) executeBeforeEachActionHooks() error {
	for _, hook := range aec.flow.beforeEachActionHooks {
		err := hook.Execute(aec)
		if err != nil {
			return fmt.Errorf("failed to execute hook before action '%s'", aec.actionName)
		}
	}
	return nil
}

func (aec *defaultActionExecutionContext) executeAfterHooks() error {
	currentStateName := aec.stash.getStateName()
	currentState, _ := aec.flow.getState(currentStateName)
	currentFlowName := currentState.getFlowName()

	var nextFlowName FlowName
	if nextStateName := aec.stash.getNextStateName(); len(nextStateName) > 0 {
		nextState, _ := aec.flow.getState(nextStateName)
		nextFlowName = nextState.getFlowName()
	}

	if len(nextFlowName) == 0 || currentFlowName != nextFlowName {
		for _, hook := range aec.flow.afterFlowHooks[currentFlowName].reverse() {
			if err := hook.Execute(aec); err != nil {
				return fmt.Errorf("failed to execute hook after flow '%s': %w", currentFlowName, err)
			}
		}
	}

	if actions := aec.flow.afterStateHooks[currentStateName]; actions != nil {
		for _, hook := range actions.reverse() {
			if err := hook.Execute(aec); err != nil {
				return fmt.Errorf("failed to execute hook action after flow '%s': %w", currentStateName, err)
			}
		}
	}

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
	return aec.inputSchema
}

// payload returns the JSONManager for accessing payload data.
func (aec *defaultActionExecutionContext) Payload() payload {
	return aec.payload
}

// CopyInputValuesToStash copies specified inputs to the stash.
func (aec *defaultActionExecutionContext) CopyInputValuesToStash(inputNames ...string) error {
	for _, inputName := range inputNames {
		// Copy input values to the stash.
		if result := aec.inputSchema.Get(inputName); result.Exists() && len(result.String()) > 0 {
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
	return aec.inputSchema.validateInputData()
}

// Error continues the flow execution to the current state, if it's a 4xx error or to the error state otherwise.
// The flow response will contain the given error.
func (aec *defaultActionExecutionContext) Error(flowErr FlowError) error {
	aec.flowError = flowErr
	statusStr := strconv.Itoa(aec.flowError.Status())
	if strings.HasPrefix(statusStr, "4") {
		if err := aec.stash.pushErrorState(aec.stash.getStateName()); err != nil {
			return err
		}
	} else {
		if err := aec.stash.pushErrorState(aec.flow.errorStateName); err != nil {
			return err
		}
	}

	return aec.closeExecutionContext()
}

// Revert reverts the flow back to the previous state.
func (aec *defaultActionExecutionContext) Revert() error {
	if err := aec.stash.revertState(); err != nil {
		return fmt.Errorf("failed to revert to the previous state: %w", err)
	}

	if err := aec.executeBeforeEachActionHooks(); err != nil {
		return err
	}

	if err := aec.executeBeforeStateHooks(aec.stash.getStateName()); err != nil {
		return err
	}

	return aec.closeExecutionContext()
}

func (aec *defaultActionExecutionContext) Continue(stateNames ...StateName) error {
	for _, stateName := range stateNames {
		if _, ok := aec.flow.stateDetails[stateName]; !ok {
			return fmt.Errorf("cannot continue, state does not exist: %s", stateName)
		}
	}

	if err := aec.executeBeforeEachActionHooks(); err != nil {
		return err
	}

	aec.stash.addScheduledStateNames(stateNames...)

	if err := aec.executeAfterHooks(); err != nil {
		return err
	}

	if err := aec.executeBeforeStateHooks(aec.stash.getNextStateName()); err != nil {
		return err
	}

	if err := aec.stash.pushState(!aec.preventRevert); err != nil {
		return fmt.Errorf("cannot continue, failed to update stash data: %s", err)
	}

	return aec.closeExecutionContext()
}

func (aec *defaultActionExecutionContext) AddLink(links ...Link) {
	aec.links = append(aec.links, links...)
}

func (aec *defaultActionExecutionContext) ScheduleStates(stateNames ...StateName) {
	aec.stash.addScheduledStateNames(stateNames...)
}

func (aec *defaultActionExecutionContext) Set(key string, value interface{}) {
	aec.flow.Set(key, value)
}

func (aec *defaultActionExecutionContext) SuspendAction() {
	aec.isSuspended = true
}

func (aec *defaultActionExecutionContext) PreventRevert() {
	aec.preventRevert = true
}
