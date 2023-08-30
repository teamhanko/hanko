package flowpilot

import (
	"time"
)

// FlowBuilder is a builder struct for creating a new Flow.
type FlowBuilder struct {
	path         string
	ttl          time.Duration
	initialState StateName
	errorState   StateName
	endState     StateName
	flow         StateTransitions
	debug        bool
}

// NewFlow creates a new FlowBuilder that builds a new flow available under the specified path.
func NewFlow(path string) *FlowBuilder {
	return &FlowBuilder{
		path: path,
		flow: make(StateTransitions),
	}
}

// TTL sets the time-to-live (TTL) for the flow.
func (fb *FlowBuilder) TTL(ttl time.Duration) *FlowBuilder {
	fb.ttl = ttl
	return fb
}

// State adds a new state transition to the FlowBuilder.
func (fb *FlowBuilder) State(state StateName, mList ...Method) *FlowBuilder {
	var transitions Transitions
	for _, m := range mList {
		transitions = append(transitions, Transition{Method: m})
	}
	fb.flow[state] = transitions
	return fb
}

// FixedStates sets the initial and final states of the flow.
func (fb *FlowBuilder) FixedStates(initialState, errorState, finalState StateName) *FlowBuilder {
	fb.initialState = initialState
	fb.errorState = errorState
	fb.endState = finalState
	return fb
}

// Debug enables the debug mode, which causes the flow response to contain the actual error.
func (fb *FlowBuilder) Debug(enabled bool) *FlowBuilder {
	fb.debug = enabled
	return fb
}

// Build constructs and returns the Flow object.
func (fb *FlowBuilder) Build() Flow {
	return Flow{
		Path:         fb.path,
		Flow:         fb.flow,
		InitialState: fb.initialState,
		ErrorState:   fb.errorState,
		EndState:     fb.endState,
		TTL:          fb.ttl,
		Debug:        fb.debug,
	}
}
