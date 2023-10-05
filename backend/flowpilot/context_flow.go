package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
)

// defaultFlowContext is the default implementation of the flowContext interface.
type defaultFlowContext struct {
	payload   utils.Payload // JSONManager for payload data.
	stash     utils.Stash   // JSONManager for stash data.
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

// GetInitialState returns the initial addState of the defaultFlow.
func (fc *defaultFlowContext) GetInitialState() StateName {
	return fc.flow.initialState
}

// GetCurrentState returns the current addState of the defaultFlow.
func (fc *defaultFlowContext) GetCurrentState() StateName {
	return fc.flowModel.CurrentState
}

// CurrentStateEquals returns true, when one of the given stateNames matches the current addState name.
func (fc *defaultFlowContext) CurrentStateEquals(stateNames ...StateName) bool {
	for _, s := range stateNames {
		if s == fc.flowModel.CurrentState {
			return true
		}
	}

	return false
}

// GetPreviousState returns a pointer to the previous addState of the flow.
func (fc *defaultFlowContext) GetPreviousState() *StateName {
	// TODO: A new state history logic needs to be implemented to reintroduce the functionality
	return nil
}

// GetErrorState returns the designated error addState of the flow.
func (fc *defaultFlowContext) GetErrorState() StateName {
	return fc.flow.errorState
}

// GetEndState returns the final addState of the flow.
func (fc *defaultFlowContext) GetEndState() StateName {
	return fc.flow.endState
}

// Stash returns the JSONManager for accessing stash data.
func (fc *defaultFlowContext) Stash() utils.Stash {
	return fc.stash
}

// StateExists checks if a given addState exists within the flow.
func (fc *defaultFlowContext) StateExists(stateName StateName) bool {
	return fc.flow.stateExists(stateName)
}

// FetchActionInput fetches input data for a specific action.
func (fc *defaultFlowContext) FetchActionInput(methodName ActionName) (utils.ReadOnlyActionInput, error) {
	// Find the last Transition with the specified method from the database wrapper.
	t, err := fc.dbw.FindLastTransitionWithAction(fc.flowModel.ID, methodName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last Transition from dbw: %w", err)
	}

	// If no Transition is found, return an empty JSONManager.
	if t == nil {
		return utils.NewActionInput(), nil
	}

	// Parse input data from the Transition.
	inputData, err := utils.NewActionInputFromString(t.InputData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Transition data: %w", err)
	}

	return inputData, nil
}
