package flowpilot

import (
	"fmt"
	"time"
)

// InputData holds input data in JSON format.
type InputData struct {
	JSONString string `json:"input_data"`
}

// getJSONStringOrDefault returns the JSON string or a default "{}" value.
func (i InputData) getJSONStringOrDefault() string {
	if len(i.JSONString) == 0 {
		return "{}"
	}

	return i.JSONString
}

// flowExecutionOptions represents options for executing a defaultFlow.
type flowExecutionOptions struct {
	action    string
	inputData InputData
}

// WithActionParam sets the ActionName for flowExecutionOptions.
func WithActionParam(action string) func(*flowExecutionOptions) {
	return func(f *flowExecutionOptions) {
		f.action = action
	}
}

// WithInputData sets the InputData for flowExecutionOptions.
func WithInputData(inputData InputData) func(*flowExecutionOptions) {
	return func(f *flowExecutionOptions) {
		f.inputData = inputData
	}
}

// StateName represents the name of a state in a defaultFlow.
type StateName string

// ActionName represents the name of an action associated with a Transition.
type ActionName string

// Action defines the interface for flow actions.
type Action interface {
	GetName() ActionName              // Get the action name.
	GetDescription() string           // Get the action description.
	Initialize(InitializationContext) // Initialize the action.
	Execute(ExecutionContext) error   // Execute the action.
}

// Actions represents a list of Action
type Actions []Action

// Transition holds an action associated with a state transition.
type Transition struct {
	Action Action
}

// getByName return the Action with the specified name.
func (a Actions) getByName(name ActionName) (Action, error) {
	for _, action := range a {
		currentName := action.GetName()

		if currentName == name {
			return action, nil
		}
	}

	return nil, fmt.Errorf("action '%s' not found", name)
}

// Transitions is a collection of Transition instances.
type Transitions []Transition

// getActions returns the Actions associated with the transition.
func (ts *Transitions) getActions() Actions {
	var actions Actions

	for _, t := range *ts {
		actions = append(actions, t.Action)
	}

	return actions
}

// stateDetail represents details for a state, including the associated flow, available sub-flows and eligible actions.
type stateDetail struct {
	flow     StateTransitions
	subFlows SubFlows
	actions  Actions
}

// stateDetails maps states to associated Actions, flows and sub-flows.
type stateDetails map[StateName]stateDetail

// StateTransitions maps states to associated Transitions.
type StateTransitions map[StateName]Transitions

// stateExists checks if a state exists in the flow.
func (st StateTransitions) stateExists(stateName StateName) bool {
	_, ok := st[stateName]
	return ok
}

// SubFlows maps a sub-flow init state to StateTransitions.
type SubFlows []SubFlow

// stateExists checks if the given state exists in a sub-flow of the current flow.
func (sfs SubFlows) stateExists(state StateName) bool {
	for _, subFlow := range sfs {
		if subFlow.getFlow().stateExists(state) {
			return true
		}
	}

	return false
}

type flow interface {
	getStateDetail(stateName StateName) (*stateDetail, error)
	getSubFlows() SubFlows
	getFlow() StateTransitions
}

// Flow represents a flow.
type Flow interface {
	Execute(db FlowDB, opts ...func(*flowExecutionOptions)) (FlowResult, error)
	ResultFromError(err error) FlowResult
	setDefaults()
	flow
}

type SubFlow interface {
	flow
}

// defaultFlow defines a flow structure with states, transitions, and settings.
type defaultFlow struct {
	flow         StateTransitions // State transitions mapping.
	subFlows     SubFlows         // The sub-flows of the current flow.
	stateDetails stateDetails     // Maps state names to flow details.
	path         string           // flow path or identifier.
	initialState StateName        // Initial state of the flow.
	errorState   StateName        // State representing errors.
	endState     StateName        // Final state of the flow.
	ttl          time.Duration    // Time-to-live for the flow.
	debug        bool             // Enables debug mode.
}

// getActionsForState returns transitions for a specified state.
func (f *defaultFlow) getStateDetail(stateName StateName) (*stateDetail, error) {
	if detail, ok := f.stateDetails[stateName]; ok {
		return &detail, nil
	}

	return nil, fmt.Errorf("unknown state: %s", stateName)
}

func (f *defaultFlow) getSubFlows() SubFlows {
	return f.subFlows
}

func (f *defaultFlow) getFlow() StateTransitions {
	return f.flow
}

// setDefaults sets default values for defaultFlow settings.
func (f *defaultFlow) setDefaults() {
	if f.ttl.Seconds() == 0 {
		f.ttl = time.Minute * 60
	}
}

// Execute handles the execution of actions for a defaultFlow.
func (f *defaultFlow) Execute(db FlowDB, opts ...func(*flowExecutionOptions)) (FlowResult, error) {
	// Process execution options.
	var executionOptions flowExecutionOptions

	for _, option := range opts {
		option(&executionOptions)
	}

	// Set default values for flow settings.
	f.setDefaults()

	if len(executionOptions.action) == 0 {
		// If the action is empty, create a new flow.
		return createAndInitializeFlow(db, *f)
	}

	// Otherwise, update an existing flow.
	return executeFlowAction(db, *f, executionOptions)
}

// ResultFromError returns an error response for the defaultFlow.
func (f *defaultFlow) ResultFromError(err error) (result FlowResult) {
	flowError := ErrorTechnical

	if e, ok := err.(FlowError); ok {
		flowError = e
	} else {
		flowError = flowError.Wrap(err)
	}

	return newFlowResultFromError(f.errorState, flowError, f.debug)
}
