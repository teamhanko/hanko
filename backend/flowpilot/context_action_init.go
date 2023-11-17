package flowpilot

// defaultActionInitializationContext is the default implementation of the actionInitializationContext interface.
type defaultActionInitializationContext struct {
	schema        InitializationSchema // InitializationSchema for action initialization.
	isSuspended   bool                 // Flag indicating if the method is suspended.
	stash         Stash                // ReadOnlyJSONManager for accessing stash data.
	contextValues contextValues        // Values to be used within the flow context.
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

// Get returns the context value with the given name.
func (aic *defaultActionInitializationContext) Get(key string) interface{} {
	return aic.contextValues[key]
}
