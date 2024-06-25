package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
)

// defaultFlowContext is the default implementation of the flowContext interface.
type defaultFlowContext struct {
	payload   payload       // JSONManager for payload data.
	stash     stash         // JSONManager for stash data.
	flow      defaultFlow   // The associated defaultFlow instance.
	dbw       flowDBWrapper // Wrapped FlowDB instance with additional functionality.
	flowModel *FlowModel    // The current FlowModel.
	csrfToken string
}

// GetFlowID returns the unique ID of the current flow.
func (fc *defaultFlowContext) GetFlowID() uuid.UUID {
	return fc.flowModel.ID
}

// GetPath returns the current path within the flow.
func (fc *defaultFlowContext) GetPath() string {
	return fc.flow.path
}

// GetInitialState returns the initial state of the flow.
func (fc *defaultFlowContext) GetInitialState() StateName {
	return fc.flow.initialStateName
}

// GetCurrentState returns the current state of the flow.
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
func (fc *defaultFlowContext) GetPreviousState() StateName {
	if fc.flowModel.PreviousState != nil {
		return *fc.flowModel.PreviousState
	}

	return ""
}

// IsPreviousState returns true if the previous state equals the given name
func (fc *defaultFlowContext) IsPreviousState(name StateName) bool {
	if fc.flowModel.PreviousState != nil {
		return *fc.flowModel.PreviousState == name
	}
	return false
}

// GetErrorState returns the designated error state of the flow.
func (fc *defaultFlowContext) GetErrorState() StateName {
	return fc.flow.errorStateName
}

// Stash returns the JSONManager for accessing stash data.
func (fc *defaultFlowContext) Stash() stash {
	return fc.stash
}

// StateExists checks if a given state exists within the current (sub-)flow.
func (fc *defaultFlowContext) StateExists(stateName StateName) bool {
	state, _ := fc.flow.getState(fc.flowModel.CurrentState)

	if state != nil {
		return state.getFlow().stateExists(stateName)
	}

	return false
}

// Get returns the context value with the given name.
func (fc *defaultFlowContext) Get(name string) interface{} {
	return fc.flow.contextValues[name]
}

// GetFlowName returns the name of the current flow.
func (fc *defaultFlowContext) GetFlowName() FlowName {
	return fc.flow.name
}

// IsFlow returns true if the name matches the current flow name.
func (fc *defaultFlowContext) IsFlow(name FlowName) bool {
	return fc.flow.name == name
}

// FetchActionInput fetches input data for a specific action.
func (fc *defaultFlowContext) FetchActionInput(methodName ActionName) (readOnlyActionInput, error) {
	// Find the last Transition with the specified method from the database wrapper.
	t, err := fc.dbw.FindLastTransitionWithAction(fc.flowModel.ID, methodName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last Transition from dbw: %w", err)
	}

	// If no Transition is found, return an empty JSONManager.
	if t == nil {
		return newActionInput(), nil
	}

	// Parse input data from the Transition.
	inputData, err := newActionInputFromString(t.InputData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Transition data: %w", err)
	}

	return inputData, nil
}
