package flowpilot

import (
	"errors"
	"fmt"
	"time"
)

type FlowBuilder interface {
	TTL(ttl time.Duration) FlowBuilder
	State(stateName StateName, actions ...Action) FlowBuilder
	FixedStates(initialStateName, errorStateName, finalStateName StateName) FlowBuilder
	Debug(enabled bool) FlowBuilder
	SubFlows(subFlows ...SubFlow) FlowBuilder
	Build() (Flow, error)
	MustBuild() Flow
}

type SubFlowBuilder interface {
	State(stateName StateName, actions ...Action) SubFlowBuilder
	SubFlows(subFlows ...SubFlow) SubFlowBuilder
	Build() (SubFlow, error)
	MustBuild() SubFlow
}

// defaultFlowBuilderBase is the base flow builder struct.
type defaultFlowBuilderBase struct {
	flow             StateTransitions
	subFlows         SubFlows
	initialStateName StateName
	stateDetails     stateDetails
}

// defaultFlowBuilder is a builder struct for creating a new Flow.
type defaultFlowBuilder struct {
	path           string
	ttl            time.Duration
	errorStateName StateName
	endStateName   StateName
	debug          bool

	defaultFlowBuilderBase
}

// newFlowBuilderBase creates a new defaultFlowBuilderBase instance.
func newFlowBuilderBase() defaultFlowBuilderBase {
	return defaultFlowBuilderBase{flow: make(StateTransitions), subFlows: make(SubFlows, 0), stateDetails: make(stateDetails)}
}

// NewFlow creates a new defaultFlowBuilder that builds a new flow available under the specified path.
func NewFlow(path string) FlowBuilder {
	fbBase := newFlowBuilderBase()

	return &defaultFlowBuilder{path: path, defaultFlowBuilderBase: fbBase}
}

// TTL sets the time-to-live (TTL) for the flow.
func (fb *defaultFlowBuilder) TTL(ttl time.Duration) FlowBuilder {
	fb.ttl = ttl
	return fb
}

func (fb *defaultFlowBuilderBase) addState(stateName StateName, actions ...Action) {
	var transitions Transitions

	for _, action := range actions {
		transitions = append(transitions, Transition{Action: action})
	}

	fb.flow[stateName] = transitions
}

func (fb *defaultFlowBuilderBase) addSubFlows(subFlows ...SubFlow) {
	fb.subFlows = append(fb.subFlows, subFlows...)
}

func (fb *defaultFlowBuilderBase) addDefaultStates(stateNames ...StateName) {
	for _, stateName := range stateNames {
		if _, exists := fb.flow[stateName]; !exists {
			fb.addState(stateName)
		}
	}
}

// State adds a new  transition to the flow.
func (fb *defaultFlowBuilder) State(stateName StateName, actions ...Action) FlowBuilder {
	fb.addState(stateName, actions...)
	return fb
}

// FixedStates sets the initial and final states of the flow.
func (fb *defaultFlowBuilder) FixedStates(initialStateName, errorStateName, finalStateName StateName) FlowBuilder {
	fb.initialStateName = initialStateName
	fb.errorStateName = errorStateName
	fb.endStateName = finalStateName
	return fb
}

func (fb *defaultFlowBuilder) SubFlows(subFlows ...SubFlow) FlowBuilder {
	fb.addSubFlows(subFlows...)
	return fb
}

// Debug enables the debug mode, which causes the flow response to contain the actual error.
func (fb *defaultFlowBuilder) Debug(enabled bool) FlowBuilder {
	fb.debug = enabled
	return fb
}

// scanFlowStates iterates through each state in the provided flow and associates relevant information, also it checks
// for uniqueness of state names.
func (fb *defaultFlowBuilder) scanFlowStates(flow StateTransitions, subFlows SubFlows) error {
	// Iterate through states in the flow.
	for stateName, transitions := range flow {
		// Check if state name is already in use.
		if _, ok := fb.stateDetails[stateName]; ok {
			return fmt.Errorf("non-unique flow state '%s'", stateName)
		}

		// Retrieve associated actions for the state.
		actions := transitions.getActions()

		// Create state details.
		state := stateDetail{
			name:     stateName,
			flow:     flow,
			subFlows: subFlows,
			actions:  actions,
		}

		// Store state details.
		fb.stateDetails[stateName] = &state
	}

	// Recursively scan sub-flows.
	for _, sf := range subFlows {
		if err := fb.scanFlowStates(sf.getFlow(), sf.getSubFlows()); err != nil {
			return err
		}
	}

	return nil
}

// validate performs validation checks on the flow configuration.
func (fb *defaultFlowBuilder) validate() error {
	// Validate fixed states and their presence in the flow.
	if len(fb.initialStateName) == 0 {
		return errors.New("fixed state 'initialState' is not set")
	}
	if len(fb.errorStateName) == 0 {
		return errors.New("fixed state 'errorState' is not set")
	}
	if len(fb.endStateName) == 0 {
		return errors.New("fixed state 'endState' is not set")
	}
	if !fb.flow.stateExists(fb.initialStateName) {
		return errors.New("fixed state 'initialState' does not belong to the flow")
	}
	if !fb.flow.stateExists(fb.errorStateName) {
		return errors.New("fixed state 'errorState' does not belong to the flow")
	}
	if !fb.flow.stateExists(fb.endStateName) {
		return errors.New("fixed state 'endState' does not belong to the flow")
	}
	if transitions, ok := fb.flow[fb.endStateName]; ok && len(transitions.getActions()) > 0 {
		return fmt.Errorf("the specified endState '%s' is not allowed to have transitions", fb.endStateName)
	}

	return nil
}

// Build constructs and returns the Flow object.
func (fb *defaultFlowBuilder) Build() (Flow, error) {
	fb.addDefaultStates(fb.initialStateName, fb.errorStateName, fb.endStateName)

	if err := fb.scanFlowStates(fb.flow, fb.subFlows); err != nil {
		return nil, fmt.Errorf("failed to scan flow states: %w", err)
	}

	if err := fb.validate(); err != nil {
		return nil, fmt.Errorf("flow validation failed: %w", err)
	}

	f := defaultFlow{
		path:             fb.path,
		flow:             fb.flow,
		initialStateName: fb.initialStateName,
		errorStateName:   fb.errorStateName,
		endStateName:     fb.endStateName,
		subFlows:         fb.subFlows,
		stateDetails:     fb.stateDetails,
		ttl:              fb.ttl,
		debug:            fb.debug,
	}

	return &f, nil
}

// MustBuild constructs and returns the Flow object, panics on error.
func (fb *defaultFlowBuilder) MustBuild() Flow {
	f, err := fb.Build()

	if err != nil {
		panic(err)
	}

	return f
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

// State adds a new  transition to the flow.
func (sfb *defaultSubFlowBuilder) State(stateName StateName, actions ...Action) SubFlowBuilder {
	sfb.addState(stateName, actions...)
	return sfb
}

// Build constructs and returns the SubFlow object.
func (sfb *defaultSubFlowBuilder) Build() (SubFlow, error) {

	f := defaultFlow{
		flow:     sfb.flow,
		subFlows: sfb.subFlows,
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
