package flowpilot

import (
	"fmt"
	"reflect"
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

// flowExecutionOptions represents options for executing a flow.
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

// StateName represents the name of a state in a flow.
type StateName string

// ActionName represents the name of an action.
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

// HookAction defines the interface for a hook action.
type HookAction interface {
	Execute(HookExecutionContext) error
}

// hookActions represents a list of HookAction interfaces.
type hookActions []HookAction

func (actions hookActions) reverse() hookActions {
	a := make(hookActions, len(actions))
	copy(a, actions)
	n := reflect.ValueOf(a).Len()
	swap := reflect.Swapper(a)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
	return a
}

// stateActions maps state names to associated actions.
type stateActions map[StateName]Actions

// stateActions maps state names to associated hook actions.
type stateHooks map[StateName]hookActions

// stateExists checks if a state exists in the flow.
func (st stateActions) stateExists(stateName StateName) bool {
	_, ok := st[stateName]
	return ok
}

// SubFlows represents a list of SubFlow interfaces.
type SubFlows []subFlow

// stateExists checks if the given state exists in a sub-flow of the current flow.
func (sfs SubFlows) stateExists(state StateName) bool {
	for _, subFlow := range sfs {
		if subFlow.getFlow().stateExists(state) {
			return true
		}
	}

	return false
}

func (sfs SubFlows) getSubFlowFromStateName(state StateName) subFlow {
	for _, subFlow := range sfs {
		if subFlow.getFlow().stateExists(state) {
			return subFlow
		}
	}
	return nil
}

// flowBase represents the base of the flow interfaces.
type flowBase interface {
	getName() string
	getSubFlows() SubFlows
	getFlow() stateActions
	getBeforeStateHooks() stateHooks
	getAfterStateHooks() stateHooks
}

// Flow represents a flow.
type Flow interface {
	// Execute executes the flow using the provided FlowDB and options.
	// It returns the result of the flow execution and an error if any.
	Execute(db FlowDB, opts ...func(*flowExecutionOptions)) (flowResult, error)
	// ResultFromError converts an error into a flowResult.
	ResultFromError(err error) flowResult
	// Set sets a value with the given key in the flow context.
	Set(string, interface{})
	// setDefaults sets the default values for the flow.
	setDefaults()
	// getState retrieves the details of a specific state in the flow.
	getState(stateName StateName) (stateDetail, error)
	// Embed the flowBase interface.
	flowBase
}

// subFlow represents a sub-flow.
type subFlow interface {
	flowBase
}

type contextValues map[string]interface{}

type defaultFlowBase struct {
	name                  string
	flow                  stateActions // StateName to Actions mapping.
	subFlows              SubFlows     // The sub-flows of the current flow.
	beforeStateHooks      stateHooks   // StateName to hookActions mapping.
	afterStateHooks       stateHooks   // StateName to hookActions mapping.
	beforeEachActionHooks hookActions  // List of hookActions that run before each action.
	afterEachActionHooks  hookActions  // List of hookActions that run after each action.
}

// defaultFlow defines a flow structure with states, actions, and settings.
type defaultFlow struct {
	stateDetails          stateDetails  // Maps state names to flow details.
	path                  string        // flow path or identifier.
	initialStateName      StateName     // Initial state of the flow.
	initialNextStateNames []StateName   // A list of next states in case a sub-flow should be invoked initially.
	errorStateName        StateName     // State representing errors.
	ttl                   time.Duration // Time-to-live for the flow.
	debug                 bool          // Enables debug mode.
	contextValues         contextValues // Values to be used within the flow context.

	defaultFlowBase
}

func (f *defaultFlow) Set(name string, value interface{}) {
	f.contextValues[name] = value
}

// getActionsForState returns state details for the specified state.
func (f *defaultFlow) getState(stateName StateName) (stateDetail, error) {
	if state, ok := f.stateDetails[stateName]; ok {
		return state, nil
	}

	return nil, fmt.Errorf("unknown state: %s", stateName)
}

// getName returns the flow name.
func (f *defaultFlowBase) getName() string {
	return f.name
}

// getSubFlows returns the sub-flows of the current flow.
func (f *defaultFlowBase) getSubFlows() SubFlows {
	return f.subFlows
}

// getFlow returns the state to action mapping of the current flow.
func (f *defaultFlowBase) getFlow() stateActions {
	return f.flow
}

func (f *defaultFlowBase) getBeforeStateHooks() stateHooks {
	return f.beforeStateHooks
}

func (f *defaultFlowBase) getAfterStateHooks() stateHooks {
	return f.afterStateHooks
}

// setDefaults sets default values for defaultFlow settings.
func (f *defaultFlow) setDefaults() {
	if f.ttl.Seconds() == 0 {
		f.ttl = time.Minute * 60
	}
}

// Execute handles the execution of actions for a defaultFlow.
func (f *defaultFlow) Execute(db FlowDB, opts ...func(*flowExecutionOptions)) (flowResult, error) {
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
func (f *defaultFlow) ResultFromError(err error) flowResult {
	flowError := ErrorTechnical

	if e, ok := err.(FlowError); ok {
		flowError = e
	} else {
		flowError = flowError.Wrap(err)
	}

	return newFlowResultFromError(f.errorStateName, flowError, f.debug)
}
