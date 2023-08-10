package flowpilot

import "hanko_flowsc/flowpilot/jsonmanager"

// defaultMethodInitializationContext is the default implementation of the methodInitializationContext interface.
type defaultMethodInitializationContext struct {
	schema      Schema                          // Schema for method initialization.
	isSuspended bool                            // Flag indicating if the method is suspended.
	stash       jsonmanager.ReadOnlyJSONManager // ReadOnlyJSONManager for accessing stash data.
}

// AddInputs adds input data to the Schema and returns a DefaultSchema instance.
func (mic *defaultMethodInitializationContext) AddInputs(inputList ...*DefaultInput) *DefaultSchema {
	return mic.schema.AddInputs(inputList...)
}

// SuspendMethod sets the isSuspended flag to indicate the method is suspended.
func (mic *defaultMethodInitializationContext) SuspendMethod() {
	mic.isSuspended = true
}

// Stash returns the ReadOnlyJSONManager for accessing stash data.
func (mic *defaultMethodInitializationContext) Stash() jsonmanager.ReadOnlyJSONManager {
	return mic.stash
}
