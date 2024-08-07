package flowpilot

// defaultActionInitializationContext is the default implementation of the actionInitializationContext interface.
type defaultActionInitializationContext struct {
	inputSchema         initializationInputSchema // initializationInputSchema for action initialization.
	isSuspended         bool                      // Flag indicating if the method is suspended.
	*defaultFlowContext                           // Embedding the defaultFlowContext for common context fields.
}

func (aic *defaultActionInitializationContext) Payload() payload {
	return aic.payload
}

func (aic *defaultActionInitializationContext) Set(s string, i interface{}) {
	aic.flow.Set(s, i)
}

// AddInputs adds input data to the initializationInputSchema.
func (aic *defaultActionInitializationContext) AddInputs(inputs ...Input) {
	aic.inputSchema.AddInputs(inputs...)
}

// SuspendAction sets the isSuspended flag to indicate the action is suspended.
func (aic *defaultActionInitializationContext) SuspendAction() {
	aic.isSuspended = true
}

// Stash returns the ReadOnlyJSONManager for accessing stash data.
func (aic *defaultActionInitializationContext) Stash() stash {
	return aic.stash
}

// Get returns the context value with the given name.
func (aic *defaultActionInitializationContext) Get(key string) interface{} {
	return aic.flow.contextValues[key]
}

func (aic *defaultActionInitializationContext) StateIsRevertible() bool {
	return aic.stash.isRevertible()
}
