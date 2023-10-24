package flowpilot

// defaultActionInitializationContext is the default implementation of the actionInitializationContext interface.
type defaultActionInitializationContext struct {
	schema      InitializationSchema // InitializationSchema for action initialization.
	isSuspended bool                 // Flag indicating if the method is suspended.
	stash       Stash                // ReadOnlyJSONManager for accessing stash data.
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
func (aic *defaultActionInitializationContext) Stash() Stash {
	return aic.stash
}

func (aic *defaultActionInitializationContext) GetCurrentState() StateName {
	return aic.currentState
}
