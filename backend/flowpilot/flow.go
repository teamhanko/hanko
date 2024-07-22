package flowpilot

import (
	"fmt"
	"reflect"
	"time"
)

// FlowName represents the name of the flow.
type FlowName string

// InputData holds input data in JSON format.
type InputData struct {
	InputDataMap map[string]interface{} `json:"input_data"`
	CSRFToken    string                 `json:"csrf_token"`
}

// WithQueryParamKey sets the ActionName for flowExecutionOptions.
func WithQueryParamKey(key string) func(*defaultFlow) {
	return func(f *defaultFlow) {
		f.queryParamKey = key
	}
}

// WithQueryParamValue sets the ActionName for flowExecutionOptions.
func WithQueryParamValue(value string) func(*defaultFlow) {
	return func(f *defaultFlow) {
		f.queryParamValue = value
	}
}

// WithInputData sets the InputData for flowExecutionOptions.
func WithInputData(inputData InputData) func(*defaultFlow) {
	return func(f *defaultFlow) {
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

func (actions hookActions) makeUnique() hookActions {
	seen := make(map[HookAction]bool)
	var uniqueSlice []HookAction

	for _, action := range actions {
		if _, found := seen[action]; !found {
			seen[action] = true
			uniqueSlice = append(uniqueSlice, action)
		}
	}

	return uniqueSlice
}

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

func (sh stateHooks) makeUnique() {
	for stateName, actions := range sh {
		sh[stateName] = actions.makeUnique()
	}
}

// flowHooks maps state names to associated hook actions.
type flowHooks map[FlowName]hookActions

func (fh flowHooks) makeUnique() {
	for stateName, actions := range fh {
		fh[stateName] = actions.makeUnique()
	}
}

// stateExists checks if a state exists in the flow.
func (st stateActions) stateExists(stateName StateName) bool {
	_, ok := st[stateName]
	return ok
}

// SubFlows represents a list of SubFlow interfaces.
type SubFlows []subFlow

// stateExists checks if the given state exists in a sub-flow of the current flow.
func (sfs SubFlows) stateExists(state StateName) bool {
	for _, sf := range sfs {
		if sf.getFlow().stateExists(state) {
			return true
		}
	}

	return false
}

func (sfs SubFlows) getSubFlowFromStateName(state StateName) subFlow {
	for _, sf := range sfs {
		if sf.getFlow().stateExists(state) {
			return sf
		}
	}
	return nil
}

// flowBase represents the base of the flow interfaces.
type flowBase interface {
	getName() FlowName
	getSubFlows() SubFlows
	getFlow() stateActions
	getBeforeStateHooks() stateHooks
	getAfterStateHooks() stateHooks
	getAfterFlowHooks() hookActions
}

// Flow represents a flow.
type Flow interface {
	// Execute executes the flow using the provided FlowDB and options.
	// It returns the result of the flow execution and an error if any.
	Execute(db FlowDB, opts ...func(*defaultFlow)) (FlowResult, error)
	// ResultFromError converts an error into a FlowResult.
	ResultFromError(err error) FlowResult
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
	name                  FlowName
	flow                  stateActions // StateName to Actions mapping.
	subFlows              SubFlows     // The sub-flows of the current flow.
	beforeStateHooks      stateHooks   // StateName to hookActions mapping.
	afterStateHooks       stateHooks   // StateName to hookActions mapping.
	beforeEachActionHooks hookActions  // List of hookActions that run before each action.
	afterEachActionHooks  hookActions  // List of hookActions that run after each action.
	afterFlowHooks        flowHooks
}

// defaultFlow defines a flow structure with states, actions, and settings.
type defaultFlow struct {
	stateDetails      stateDetails  // Maps state names to flow details.
	initialStateNames []StateName   // A list of next states in case a sub-flow should be invoked initially.
	errorStateName    StateName     // State representing errors.
	ttl               time.Duration // Time-to-live for the flow.
	debug             bool          // Enables debug mode.
	queryParam        queryParam    // TODO
	contextValues     contextValues // Values to be used within the flow context.
	inputData         InputData
	queryParamKey     string
	queryParamValue   string

	*defaultFlowBase
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
func (f *defaultFlowBase) getName() FlowName {
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

func (f *defaultFlowBase) getAfterFlowHooks() hookActions {
	return f.afterFlowHooks[f.name]
}

// setDefaults sets default values for defaultFlow settings.
func (f *defaultFlow) setDefaults() {
	if f.ttl.Seconds() == 0 {
		f.ttl = time.Minute * 60
	}

	if len(f.queryParamKey) == 0 {
		f.queryParamKey = "flowpilot_action"
	}
}

// Execute handles the execution of actions for a defaultFlow.
func (f *defaultFlow) Execute(db FlowDB, opts ...func(flow *defaultFlow)) (FlowResult, error) {
	for _, option := range opts {
		option(f)
	}

	// Set default values for flow settings.
	f.setDefaults()

	// If the action is empty, create a new flow.
	if len(f.queryParamValue) == 0 {
		return createAndInitializeFlow(db, *f)
	}

	// Otherwise, update an existing flow...
	q, err := newQueryParam(f.queryParamKey, f.queryParamValue)
	if err != nil {
		return newFlowResultFromError(f.errorStateName, ErrorTechnical.Wrap(err), f.debug), nil
	}

	f.queryParam = q

	return executeFlowAction(db, *f)
}

// ResultFromError returns an error response for the defaultFlow.
func (f *defaultFlow) ResultFromError(err error) FlowResult {
	flowError := ErrorTechnical

	if e, ok := err.(FlowError); ok {
		flowError = e
	} else {
		flowError = flowError.Wrap(err)
	}

	return newFlowResultFromError(f.errorStateName, flowError, f.debug)
}
