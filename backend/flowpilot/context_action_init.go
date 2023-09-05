package flowpilot

import "github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"

// defaultActionInitializationContext is the default implementation of the actionInitializationContext interface.
type defaultActionInitializationContext struct {
	schema      InitializationSchema            // InitializationSchema for action initialization.
	isSuspended bool                            // Flag indicating if the method is suspended.
	stash       jsonmanager.ReadOnlyJSONManager // ReadOnlyJSONManager for accessing stash data.
}

// AddInputs adds input data to the InitializationSchema and returns a defaultSchema instance.
func (mic *defaultActionInitializationContext) AddInputs(inputList ...Input) {
	mic.schema.AddInputs(inputList...)
}

// SuspendAction sets the isSuspended flag to indicate the action is suspended.
func (mic *defaultActionInitializationContext) SuspendAction() {
	mic.isSuspended = true
}

// Stash returns the ReadOnlyJSONManager for accessing stash data.
func (mic *defaultActionInitializationContext) Stash() jsonmanager.ReadOnlyJSONManager {
	return mic.stash
}
