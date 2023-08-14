package flowpilot

import "github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"

// defaultMethodInitializationContext is the default implementation of the methodInitializationContext interface.
type defaultMethodInitializationContext struct {
	schema      InitializationSchema            // InitializationSchema for method initialization.
	isSuspended bool                            // Flag indicating if the method is suspended.
	stash       jsonmanager.ReadOnlyJSONManager // ReadOnlyJSONManager for accessing stash data.
}

// AddInputs adds input data to the InitializationSchema and returns a defaultSchema instance.
func (mic *defaultMethodInitializationContext) AddInputs(inputList ...Input) {
	mic.schema.AddInputs(inputList...)
}

// SuspendMethod sets the isSuspended flag to indicate the method is suspended.
func (mic *defaultMethodInitializationContext) SuspendMethod() {
	mic.isSuspended = true
}

// Stash returns the ReadOnlyJSONManager for accessing stash data.
func (mic *defaultMethodInitializationContext) Stash() jsonmanager.ReadOnlyJSONManager {
	return mic.stash
}
