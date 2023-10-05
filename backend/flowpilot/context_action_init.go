package flowpilot

import (
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
)

// defaultActionInitializationContext is the default implementation of the actionInitializationContext interface.
type defaultActionInitializationContext struct {
	schema      InitializationSchema // InitializationSchema for action initialization.
	isSuspended bool                 // Flag indicating if the method is suspended.
	stash       utils.Stash          // ReadOnlyJSONManager for accessing stash data.
}

// AddInputs adds input data to the InitializationSchema.
func (aic *defaultActionInitializationContext) AddInputs(inputs ...Input) {
	aic.schema.AddInputs(inputs...)
}

// SuspendAction sets the isSuspended flag to indicate the action is suspended.
func (aic *defaultActionInitializationContext) SuspendAction() {
	aic.isSuspended = true
}

// Stash returns the ReadOnlyJSONManager for accessing stash data.
func (aic *defaultActionInitializationContext) Stash() utils.Stash {
	return aic.stash
}
