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

// HookActions represents a list of HookAction interfaces.
type HookActions []HookAction

// state represents details for a state, including the associated actions, available sub-flows and more.
type stateDetail struct {
	name        StateName
	flow        stateActions
	subFlows    SubFlows
	actions     Actions
	beforeHooks HookActions
}

// getAction returns the Action with the specified name.
func (sd *stateDetail) getAction(actionName ActionName) (Action, error) {
	for _, action := range sd.actions {
		currentActionName := action.GetName()

		if currentActionName == actionName {
			return action, nil
		}
	}

	return nil, fmt.Errorf("action '%s' not found", actionName)
}

// stateDetails maps states to associated Actions, flows and sub-flows.
type stateDetails map[StateName]*stateDetail

// stateActions maps state names to associated actions.
type stateActions map[StateName]Actions

// stateActions maps state names to associated hook actions.
type stateHooks map[StateName]HookActions

// stateExists checks if a state exists in the flow.
func (st stateActions) stateExists(stateName StateName) bool {
	_, ok := st[stateName]
	return ok
}

// SubFlows represents a list of SubFlow interfaces.
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

// flowBase represents the base of the flow interfaces.
type flowBase interface {
	getState(stateName StateName) (*stateDetail, error)
	getSubFlows() SubFlows
	getFlow() stateActions
	getBeforeHooks() stateHooks
}

// Flow represents a flow.
type Flow interface {
	Execute(db FlowDB, opts ...func(*flowExecutionOptions)) (FlowResult, error)
	ResultFromError(err error) FlowResult
	setDefaults()

	flowBase
}

// SubFlow represents a sub-flow.
type SubFlow interface {
	flowBase
}

// defaultFlow defines a flow structure with states, actions, and settings.
type defaultFlow struct {
	flow             stateActions  // StateName to Actions mapping.
	subFlows         SubFlows      // The sub-flows of the current flow.
	stateDetails     stateDetails  // Maps state names to flow details.
	path             string        // flow path or identifier.
	beforeHooks      stateHooks    // StateName to HookActions mapping.
	initialStateName StateName     // Initial state of the flow.
	errorStateName   StateName     // State representing errors.
	endStateName     StateName     // Final state of the flow.
	ttl              time.Duration // Time-to-live for the flow.
	debug            bool          // Enables debug mode.
}

// getActionsForState returns state details for the specified state.
func (f *defaultFlow) getState(stateName StateName) (*stateDetail, error) {
	if state, ok := f.stateDetails[stateName]; ok {
		return state, nil
	}

	return nil, fmt.Errorf("unknown state: %s", stateName)
}

// getSubFlows returns the sub-flows of the current flow.
func (f *defaultFlow) getSubFlows() SubFlows {
	return f.subFlows
}

// getFlow returns the state to action mapping of the current flow.
func (f *defaultFlow) getFlow() stateActions {
	return f.flow
}

func (f *defaultFlow) getBeforeHooks() stateHooks {
	return f.beforeHooks
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
func (f *defaultFlow) ResultFromError(err error) FlowResult {
	flowError := ErrorTechnical

	if e, ok := err.(FlowError); ok {
		flowError = e
	} else {
		flowError = flowError.Wrap(err)
	}

	return newFlowResultFromError(f.errorStateName, flowError, f.debug)
}
