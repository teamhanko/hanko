package flowpilot

import (
	"github.com/gofrs/uuid"
)

// defaultFlowContext is the default implementation of the flowContext interface.
type defaultFlowContext struct {
	payload   payload       // JSONManager for payload data.
	stash     stash         // JSONManager for stash data.
	flow      defaultFlow   // The associated defaultFlow instance.
	dbw       flowDBWrapper // Wrapped FlowDB instance with additional functionality.
	flowModel *FlowModel    // The current FlowModel.
}

// GetFlowID returns the unique ID of the current flow.
func (fc *defaultFlowContext) GetFlowID() uuid.UUID {
	return fc.flowModel.ID
}

// GetInitialState returns the initial state of the flow.
func (fc *defaultFlowContext) GetInitialState() StateName {
	return fc.flow.initialStateNames[0]
}

// GetCurrentState returns the current state of the flow.
func (fc *defaultFlowContext) GetCurrentState() StateName {
	return fc.stash.getStateName()
}

func (fc *defaultFlowContext) GetScheduledStates() []StateName {
	return fc.stash.getScheduledStateNames()
}

// CurrentStateEquals returns true, when one of the given stateNames matches the current state name.
func (fc *defaultFlowContext) CurrentStateEquals(stateNames ...StateName) bool {
	for _, s := range stateNames {
		if s == fc.stash.getStateName() {
			return true
		}
	}

	return false
}

func (fc *defaultFlowContext) IsStateScheduled(name StateName) bool {
	for _, state := range fc.stash.getScheduledStateNames() {
		if state == name {
			return true
		}
	}
	return false
}

func (fc *defaultFlowContext) StateVisited(name StateName) bool {
	return fc.stash.stateVisited(name)
}

// GetPreviousState returns the previous state of the flow.
func (fc *defaultFlowContext) GetPreviousState() StateName {
	return fc.stash.getPreviousStateName()
}

// IsPreviousState returns true if the previous state equals the given name
func (fc *defaultFlowContext) IsPreviousState(name StateName) bool {
	return fc.stash.getPreviousStateName() == name
}

// GetErrorState returns the designated error state of the flow.
func (fc *defaultFlowContext) GetErrorState() StateName {
	return fc.flow.errorStateName
}

// Stash returns the JSONManager for accessing stash data.
func (fc *defaultFlowContext) Stash() stash {
	return fc.stash
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
