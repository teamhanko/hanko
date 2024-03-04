package flowpilot

// defaultActionFinalizationContext is the default implementation of the actionFinalizationContext interface.
type defaultActionFinalizationContext struct {
	executionResult *executionResult // Result of the action execution.
	contextValues   contextValues
}

func (afc *defaultActionFinalizationContext) Get(key string) interface{} {
	return afc.contextValues[key]
}

func (afc *defaultActionFinalizationContext) SuspendAction() {
	afc.executionResult.isSuspended = true
}
