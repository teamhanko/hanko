package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

// defaultFlowContext is the default implementation of the flowContext interface.
type defaultFlowContext struct {
	payload   jsonmanager.JSONManager // JSONManager for payload data.
	stash     jsonmanager.JSONManager // JSONManager for stash data.
	flow      Flow                    // The associated Flow instance.
	dbw       FlowDBWrapper           // Wrapped FlowDB instance with additional functionality.
	flowModel FlowModel               // The current FlowModel.
}

// GetFlowID returns the unique ID of the current Flow.
func (fc *defaultFlowContext) GetFlowID() uuid.UUID {
	return fc.flowModel.ID
}

// GetPath returns the current path within the Flow.
func (fc *defaultFlowContext) GetPath() string {
	return fc.flow.Path
}

// GetInitialState returns the initial state of the Flow.
func (fc *defaultFlowContext) GetInitialState() StateName {
	return fc.flow.InitialState
}

// GetCurrentState returns the current state of the Flow.
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

// GetPreviousState returns a pointer to the previous state of the Flow.
func (fc *defaultFlowContext) GetPreviousState() *StateName {
	return fc.flowModel.PreviousState
}

// GetErrorState returns the designated error state of the Flow.
func (fc *defaultFlowContext) GetErrorState() StateName {
	return fc.flow.ErrorState
}

// GetEndState returns the final state of the Flow.
func (fc *defaultFlowContext) GetEndState() StateName {
	return fc.flow.EndState
}

// Stash returns the JSONManager for accessing stash data.
func (fc *defaultFlowContext) Stash() jsonmanager.JSONManager {
	return fc.stash
}

// StateExists checks if a given state exists within the Flow.
func (fc *defaultFlowContext) StateExists(stateName StateName) bool {
	return fc.flow.stateExists(stateName)
}

// FetchMethodInput fetches input data for a specific method.
func (fc *defaultFlowContext) FetchMethodInput(methodName MethodName) (jsonmanager.ReadOnlyJSONManager, error) {
	// Find the last Transition with the specified method from the database wrapper.
	t, err := fc.dbw.FindLastTransitionWithMethod(fc.flowModel.ID, methodName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last Transition from dbw: %w", err)
	}

	// If no Transition is found, return an empty JSONManager.
	if t == nil {
		return jsonmanager.NewJSONManager(), nil
	}

	// Parse input data from the Transition.
	inputData, err := jsonmanager.NewJSONManagerFromString(t.InputData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Transition data: %w", err)
	}

	return inputData, nil
}

// getCurrentTransitions retrieves the Transitions for the current state.
func (fc *defaultFlowContext) getCurrentTransitions() *Transitions {
	return fc.flow.getTransitionsForState(fc.flowModel.CurrentState)
}
