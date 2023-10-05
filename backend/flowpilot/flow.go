package flowpilot

import (
	"errors"
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

// TODO: Should it be possible to partially implement the Action interface? E.g. when a action does not require initialization.

// Action defines the interface for flow actions.
type Action interface {
	GetName() ActionName              // Get the action name.
	GetDescription() string           // Get the action description.
	Initialize(InitializationContext) // Initialize the action.
	Execute(ExecutionContext) error   // Execute the action.
}

type Actions []Action

// Transition holds an action associated with a state transition.
type Transition struct {
	Action Action
}

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

// getActions returns the Action associated with the specified name.
func (ts *Transitions) getActions() []Action {
	var actions []Action

	for _, t := range *ts {
		actions = append(actions, t.Action)
	}

	return actions
}

type stateDetail struct {
	flow     StateTransitions
	subFlows SubFlows
	actions  Actions
}

// stateDetails maps states to associated Actions, flows and sub-flows.
type stateDetails map[StateName]stateDetail

// StateTransitions maps states to associated Transitions.
type StateTransitions map[StateName]Transitions

// SubFlows maps a sub-flow init state to StateTransitions.
type SubFlows map[StateName]SubFlow

func (sfs SubFlows) isEntryStateAllowed(entryState StateName) bool {
	for _, subFlow := range sfs {
		subFlowInitState := subFlow.getInitialState()

		if subFlowInitState == entryState {
			return true
		}
	}

	return false
}

type flow interface {
	stateExists(stateName StateName) bool
	getStateDetail(stateName StateName) (*stateDetail, error)
	getSubFlows() SubFlows
	getFlow() StateTransitions
}

type Flow interface {
	Execute(db FlowDB, opts ...func(*flowExecutionOptions)) (FlowResult, error)
	ResultFromError(err error) FlowResult
	setDefaults()
	validate() error
	flow
}

type SubFlow interface {
	getInitialState() StateName
	flow
}

// defaultFlow defines a flow structure with states, transitions, and settings.
type defaultFlow struct {
	flow         StateTransitions // State transitions mapping.
	subFlows     SubFlows         // TODO
	stateDetails stateDetails     //
	path         string           // flow path or identifier.
	initialState StateName        // Initial state of the flow.
	errorState   StateName        // State representing errors.
	endState     StateName        // Final state of the flow.
	ttl          time.Duration    // Time-to-live for the flow.
	debug        bool             // Enables debug mode.
}

// stateExists checks if a state exists in the defaultFlow.
func (f *defaultFlow) stateExists(stateName StateName) bool {
	_, ok := f.flow[stateName]
	return ok
}

// getActionsForState returns transitions for a specified state.
func (f *defaultFlow) getStateDetail(stateName StateName) (*stateDetail, error) {
	if detail, ok := f.stateDetails[stateName]; ok {
		return &detail, nil
	}

	return nil, fmt.Errorf("unknown state: %s", stateName)
}

func (f *defaultFlow) getInitialState() StateName {
	return f.initialState
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

// TODO: validate while building the flow
// validate performs validation checks on the defaultFlow configuration.
func (f *defaultFlow) validate() error {
	// Validate fixed states and their presence in the flow.
	if len(f.initialState) == 0 {
		return errors.New("fixed state 'initialState' is not set")
	}
	if len(f.errorState) == 0 {
		return errors.New("fixed state 'errorState' is not set")
	}
	if len(f.endState) == 0 {
		return errors.New("fixed state 'endState' is not set")
	}
	if !f.stateExists(f.initialState) {
		return errors.New("fixed state 'initialState' does not belong to the flow")
	}
	if !f.stateExists(f.errorState) {
		return errors.New("fixed state 'errorState' does not belong to the flow")
	}
	if !f.stateExists(f.endState) {
		return errors.New("fixed state 'endState' does not belong to the flow")
	}
	if detail, _ := f.getStateDetail(f.endState); detail == nil || len(detail.actions) > 0 {
		return fmt.Errorf("the specified endState '%s' is not allowed to have transitions", f.endState)
	}

	return nil
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

	// Perform validation checks on the flow configuration.
	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("invalid flow: %w", err)
	}

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
