package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
)

// defaultFlowContext is the default implementation of the flowContext interface.
type defaultFlowContext struct {
	payload   Payload       // JSONManager for payload data.
	stash     Stash         // JSONManager for stash data.
	flow      defaultFlow   // The associated defaultFlow instance.
	dbw       FlowDBWrapper // Wrapped FlowDB instance with additional functionality.
	flowModel FlowModel     // The current FlowModel.
}

// GetFlowID returns the unique ID of the current defaultFlow.
func (fc *defaultFlowContext) GetFlowID() uuid.UUID {
	return fc.flowModel.ID
}

// GetPath returns the current path within the defaultFlow.
func (fc *defaultFlowContext) GetPath() string {
	return fc.flow.path
}

// GetInitialState returns the initial state of the defaultFlow.
func (fc *defaultFlowContext) GetInitialState() StateName {
	return fc.flow.initialState
}

// GetCurrentState returns the current state of the defaultFlow.
func (fc *defaultFlowContext) GetCurrentState() StateName {
	return fc.flowModel.CurrentState
}

// CurrentStateEquals returns true, when one of the given stateNames matches the current state name.
func (fc *defaultFlowContext) CurrentStateEquals(stateNames ...StateName) bool {
	for _, s := range stateNames {
		if s == fc.flowModel.CurrentState {
			return true
		}
	}

	return false
}

// GetPreviousState returns a pointer to the previous state of the flow.
func (fc *defaultFlowContext) GetPreviousState() *StateName {
	state, _, _, _ := fc.stash.getLastStateFromHistory()

	if state == nil {
		state = &fc.flow.initialState
	}

	return state
}

// GetErrorState returns the designated error state of the flow.
func (fc *defaultFlowContext) GetErrorState() StateName {
	return fc.flow.errorState
}

// GetEndState returns the final state of the flow.
func (fc *defaultFlowContext) GetEndState() StateName {
	return fc.flow.endState
}

// Stash returns the JSONManager for accessing stash data.
func (fc *defaultFlowContext) Stash() Stash {
	return fc.stash
}

// StateExists checks if a given state exists within the current (sub-)flow.
func (fc *defaultFlowContext) StateExists(stateName StateName) bool {
	detail, _ := fc.flow.getStateDetail(fc.flowModel.CurrentState)

	if detail != nil {
		return detail.flow.stateExists(stateName)
	}

	return false
}

// FetchActionInput fetches input data for a specific action.
func (fc *defaultFlowContext) FetchActionInput(methodName ActionName) (ReadOnlyActionInput, error) {
	// Find the last Transition with the specified method from the database wrapper.
	t, err := fc.dbw.FindLastTransitionWithAction(fc.flowModel.ID, methodName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last Transition from dbw: %w", err)
	}

	// If no Transition is found, return an empty JSONManager.
	if t == nil {
		return NewActionInput(), nil
	}

	// Parse input data from the Transition.
	inputData, err := NewActionInputFromString(t.InputData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Transition data: %w", err)
	}

	return inputData, nil
}
