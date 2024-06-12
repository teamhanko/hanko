package flowpilot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

type context interface {
	// Get returns the context value with the given name.
	Get(string) interface{}
}

// flowContext represents the basic context for a flow.
type flowContext interface {
	// GetFlowID returns the unique ID of the current defaultFlow.
	GetFlowID() uuid.UUID
	// GetPath returns the current path within the flow.
	GetPath() string

	GetFlowPath() flowPath
	// Payload returns the JSONManager for accessing payload data.
	Payload() Payload
	// Stash returns the JSONManager for accessing stash data.
	Stash() Stash
	// GetInitialState returns the initial state of the flow.
	GetInitialState() StateName
	// GetCurrentState returns the current state of the flow.
	GetCurrentState() StateName
	// CurrentStateEquals returns true, when one of the given states matches the current state.
	CurrentStateEquals(stateNames ...StateName) bool
	// GetPreviousState returns the previous state of the flow.
	GetPreviousState() (*StateName, error)
	// GetErrorState returns the designated error state of the flow.
	GetErrorState() StateName
	// StateExists checks if a given state exists within the flow.
	StateExists(stateName StateName) bool
	// Set sets a context value for the given key.
	Set(string, interface{})

	GetFlowName() string
}

// actionInitializationContext represents the basic context for a flow action's initialization.
type actionInitializationContext interface {
	// AddInputs adds input parameters to the schema.
	AddInputs(inputs ...Input)
	// Stash returns the ReadOnlyJSONManager for accessing stash data.
	Stash() Stash
	// CurrentStateEquals returns true, when one of the given states matches the current state.
	CurrentStateEquals(stateNames ...StateName) bool
	actionSuspender
	flowContext
}

// actionExecutionContext represents the context for an action execution.
type actionExecutionContext interface {
	actionSuspender
	flowContext
	// Input returns the ExecutionSchema for the action.
	Input() ExecutionSchema

	// TODO: FetchActionInput (for a action name) is maybe useless and can be removed or replaced.

	// FetchActionInput fetches input data for a specific action.
	FetchActionInput(actionName ActionName) (ReadOnlyActionInput, error)
	// ValidateInputData validates the input data against the schema.
	ValidateInputData() bool
	// CopyInputValuesToStash copies specified inputs to the stash.
	CopyInputValuesToStash(inputNames ...string) error
}

// actionExecutionContinuationContext represents the context within an action continuation.
type actionExecutionContinuationContext interface {
	actionExecutionContext
	// ContinueFlow continues the flow execution to the specified next state.
	ContinueFlow(nextStateName StateName) error
	// ContinueFlowWithError continues the flow execution to the specified next state with an error.
	ContinueFlowWithError(nextStateName StateName, flowErr FlowError) error
	// StartSubFlow starts a sub-flow and continues the flow execution to the specified next states after the sub-flow has been ended.
	StartSubFlow(initStateName StateName, nextStateNames ...StateName) error
	// EndSubFlow ends the sub-flow and continues the flow execution to the previously specified next states.
	EndSubFlow() error
	// ContinueToPreviousState rewinds the flow back to the previous state.
	ContinueToPreviousState() error
}

type actionSuspender interface {
	// SuspendAction suspends the current action's execution.
	SuspendAction()
}

type actionFinalizationContext interface {
	actionSuspender
}

type Context interface {
	context
}

// InitializationContext is a shorthand for actionInitializationContext within the flow initialization method.
type InitializationContext interface {
	context
	actionInitializationContext
}

// ExecutionContext is a shorthand for actionExecutionContinuationContext within flow execution method.
type ExecutionContext interface {
	context
	actionExecutionContinuationContext
}

type FinalizationContext interface {
	context
	actionFinalizationContext
}

type HookExecutionContext interface {
	context
	actionExecutionContext
	SetFlowError(FlowError)
	GetFlowError() FlowError
	AddLink(...Link)
}

type BeforeEachActionExecutionContext interface {
	actionExecutionContinuationContext
}

// TODO: The following interfaces are meant for a plugin system. #tbd

// PluginBeforeActionExecutionContext represents the context for a plugin before an action execution.
type PluginBeforeActionExecutionContext interface {
	actionExecutionContinuationContext
}

// PluginAfterActionExecutionContext represents the context for a plugin after an action execution.
type PluginAfterActionExecutionContext interface {
	actionExecutionContext
}

// createAndInitializeFlow initializes the flow and returns a flow Response.
func createAndInitializeFlow(db FlowDB, flow defaultFlow) (FlowResult, error) {
	// Wrap the provided FlowDB with additional functionality.
	dbw := wrapDB(db)
	// Calculate the expiration time for the flow.
	expiresAt := time.Now().Add(flow.ttl).UTC()

	// Initialize JSONManagers for stash and payload.
	stash := NewStash()
	payload := NewPayload()

	// Add the initial next states to the stash to be scheduled after the initial sub-flow.
	err := stash.addScheduledStates(flow.initialNextStateNames...)
	if err != nil {
		return nil, fmt.Errorf("failed to stash scheduled states: %w", err)
	}

	flowPath := newFlowPathFromString(flow.name)

	subflow := flow.subFlows.getSubFlowFromStateName(flow.initialStateName)
	if subflow != nil {
		flowPath.add(subflow.getName())
	}

	err = stash.Set("_.flowPath", flowPath.String())
	if err != nil {
		return nil, fmt.Errorf("failed to stash current flowPath: %w", err)
	}

	// Create a new flow model with the provided parameters.
	flowCreation := flowCreationParam{currentState: flow.initialStateName, stash: stash.String(), expiresAt: expiresAt}
	flowModel, err := dbw.createFlowWithParam(flowCreation)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow: %w", err)
	}

	// Create a defaultFlowContext instance.
	fc := defaultFlowContext{
		flow:      flow,
		dbw:       dbw,
		flowModel: *flowModel,
		stash:     stash,
		payload:   payload,
	}

	er := executionResult{nextStateName: flowModel.CurrentState}

	aec := defaultActionExecutionContext{
		actionName:         "",
		input:              nil,
		executionResult:    &er,
		defaultFlowContext: fc,
	}

	err = aec.executeBeforeStateHooks(flow.initialStateName)
	if err != nil {
		return nil, fmt.Errorf("failed to execute before hook actions: %w", err)
	}

	return er.generateResponse(fc, flow.debug), nil
}

// executeFlowAction processes the flow and returns a Response.
func executeFlowAction(db FlowDB, flow defaultFlow, options flowExecutionOptions) (FlowResult, error) {
	// Parse the actionParam parameter to get the actionParam name and flow ID.
	actionParam, err := parseActionParam(options.action)
	if err != nil {
		return newFlowResultFromError(flow.errorStateName, ErrorActionParamInvalid.Wrap(err), flow.debug), nil
	}

	// Retrieve the flow model from the database using the flow ID.
	flowModel, err := db.GetFlow(actionParam.flowID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newFlowResultFromError(flow.errorStateName, ErrorOperationNotPermitted.Wrap(err), flow.debug), nil
		}
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	// Check if the flow has expired.
	if time.Now().After(flowModel.ExpiresAt) {
		return newFlowResultFromError(flow.errorStateName, ErrorFlowExpired, flow.debug), nil
	}

	// Parse stash data from the flow model.
	stash, err := NewStashFromString(flowModel.StashData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stash from flow: %w", err)
	}

	// Initialize JSONManagers for payload and flash data.
	payload := NewPayload()

	// Create a defaultFlowContext instance.
	fc := defaultFlowContext{
		flow:      flow,
		dbw:       wrapDB(db),
		flowModel: *flowModel,
		stash:     stash,
		payload:   payload,
	}

	state, err := flow.getState(flowModel.CurrentState)
	if err != nil {
		return nil, err
	}

	// Parse raw input data into JSONManager.
	raw := options.inputData.getJSONStringOrDefault()
	inputJSON, err := NewActionInputFromString(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input data: %w", err)
	}

	// Create a ActionName from the parsed actionParam name.
	actionName := ActionName(actionParam.actionName)

	// Get the action associated with the actionParam name.
	actionDetail, err := state.getActionDetail(actionName)
	if err != nil {
		return newFlowResultFromError(flow.errorStateName, ErrorOperationNotPermitted.Wrap(err), flow.debug), nil
	}

	// Initialize the schema and action context for action execution.
	schema := newSchemaWithInputData(inputJSON)
	aic := defaultActionInitializationContext{
		schema:             schema.toInitializationSchema(),
		defaultFlowContext: fc,
	}

	// Create a actionExecutionContext instance for action execution.
	aec := defaultActionExecutionContext{
		actionName:         actionName,
		input:              schema,
		defaultFlowContext: fc,
	}

	err = aec.executeBeforeEachActionHooks()
	if err != nil {
		return newFlowResultFromError(flow.errorStateName, ErrorOperationNotPermitted, flow.debug), nil
	}

	actionDetail.action.Initialize(&aic)

	// Check if the action is suspended.
	if aic.isSuspended {
		return newFlowResultFromError(flow.errorStateName, ErrorOperationNotPermitted, flow.debug), nil
	}

	// Execute the action and handle any errors.
	err = actionDetail.action.Execute(&aec)
	if err != nil {
		return nil, fmt.Errorf("the action failed to handle the request: %w", err)
	}

	// Ensure that the action has set a result object.
	if aec.executionResult == nil {
		er := executionResult{nextStateName: flowModel.CurrentState}
		aec.executionResult = &er
	}

	// Generate a response based on the execution result.
	return aec.executionResult.generateResponse(fc, flow.debug), nil
}
