package flowpilot

type SubFlowBuilder interface {
	State(stateName StateName, actions ...Action) SubFlowBuilder
	BeforeState(stateName StateName, hooks ...HookAction) SubFlowBuilder
	AfterState(stateName StateName, hooks ...HookAction) SubFlowBuilder
	SubFlows(subFlows ...SubFlow) SubFlowBuilder
	Build() (SubFlow, error)
	MustBuild() SubFlow
}

// defaultFlowBuilder is a builder struct for creating a new SubFlow.
type defaultSubFlowBuilder struct {
	defaultFlowBuilderBase
}

// NewSubFlow creates a new SubFlowBuilder.
func NewSubFlow() SubFlowBuilder {
	fbBase := newFlowBuilderBase()
	return &defaultSubFlowBuilder{defaultFlowBuilderBase: fbBase}
}

func (sfb *defaultSubFlowBuilder) SubFlows(subFlows ...SubFlow) SubFlowBuilder {
	sfb.addSubFlows(subFlows...)
	return sfb
}

// State adds a new state to the flow.
func (sfb *defaultSubFlowBuilder) State(stateName StateName, actions ...Action) SubFlowBuilder {
	sfb.addState(stateName, actions...)
	return sfb
}

func (sfb *defaultSubFlowBuilder) BeforeState(stateName StateName, hooks ...HookAction) SubFlowBuilder {
	sfb.addBeforeStateHooks(stateName, hooks...)
	return sfb
}

func (sfb *defaultSubFlowBuilder) AfterState(stateName StateName, hooks ...HookAction) SubFlowBuilder {
	sfb.addAfterStateHooks(stateName, hooks...)
	return sfb
}

// Build constructs and returns the SubFlow object.
func (sfb *defaultSubFlowBuilder) Build() (SubFlow, error) {

	f := defaultFlowBase{
		flow:             sfb.flow,
		subFlows:         sfb.subFlows,
		beforeStateHooks: sfb.beforeStateHooks,
		afterStateHooks:  sfb.afterStateHooks,
	}

	return &f, nil
}

// MustBuild constructs and returns the SubFlow object, panics on error.
func (sfb *defaultSubFlowBuilder) MustBuild() SubFlow {
	sf, err := sfb.Build()

	if err != nil {
		panic(err)
	}

	return sf
}
