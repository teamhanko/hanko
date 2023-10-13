package flowpilot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
	"time"
)

// flowContext represents the basic context for a flow.
type flowContext interface {
	// GetFlowID returns the unique ID of the current defaultFlow.
	GetFlowID() uuid.UUID
	// GetPath returns the current path within the flow.
	GetPath() string
	// Payload returns the JSONManager for accessing payload data.
	Payload() Payload
	// Stash returns the JSONManager for accessing stash data.
	Stash() Stash
	// GetInitialState returns the initial state of the flow.
	GetInitialState() StateName
	// GetCurrentState returns the current state of the flow.
	GetCurrentState() StateName
	// CurrentStateEquals returns true, when one of the given states matches the current state.
	CurrentStateEquals(states ...StateName) bool
	// GetPreviousState returns the previous state of the flow.
	GetPreviousState() *StateName
	// GetErrorState returns the designated error state of the flow.
	GetErrorState() StateName
	// GetEndState returns the final state of the flow.
	GetEndState() StateName
	// StateExists checks if a given state exists within the flow.
	StateExists(stateName StateName) bool
}

// actionInitializationContext represents the basic context for a flow action's initialization.
type actionInitializationContext interface {
	// AddInputs adds input parameters to the schema.
	AddInputs(inputs ...Input)
	// Stash returns the ReadOnlyJSONManager for accessing stash data.
	Stash() Stash
	// SuspendAction suspends the current action's execution.
	SuspendAction()
}

// actionExecutionContext represents the context for an action execution.
type actionExecutionContext interface {
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
	ContinueFlow(nextState StateName) error
	// ContinueFlowWithError continues the flow execution to the specified next state with an error.
	ContinueFlowWithError(nextState StateName, flowErr FlowError) error
	// StartSubFlow starts a sub-flow and continues the flow execution to the specified next states after the sub-flow has been ended.
	StartSubFlow(initState StateName, nextStates ...StateName) error
	// EndSubFlow ends the sub-flow and continues the flow execution to the previously specified next states.
	EndSubFlow() error
	// ContinueToPreviousState rewinds the flow back to the previous state.
	ContinueToPreviousState() error
}

// InitializationContext is a shorthand for actionInitializationContext within the flow initialization method.
type InitializationContext interface {
	actionInitializationContext
}

// ExecutionContext is a shorthand for actionExecutionContinuationContext within flow execution method.
type ExecutionContext interface {
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

	// Create a new flow model with the provided parameters.
	flowCreation := flowCreationParam{currentState: flow.initialState, expiresAt: expiresAt}
	flowModel, err := dbw.CreateFlowWithParam(flowCreation)
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

	// Generate a response based on the execution result.
	er := executionResult{nextState: flowModel.CurrentState}

	return er.generateResponse(fc, flow.debug), nil
}

// executeFlowAction processes the flow and returns a Response.
func executeFlowAction(db FlowDB, flow defaultFlow, options flowExecutionOptions) (FlowResult, error) {
	// Parse the actionParam parameter to get the actionParam name and flow ID.
	actionParam, err := utils.ParseActionParam(options.action)
	if err != nil {
		return newFlowResultFromError(flow.errorState, ErrorActionParamInvalid.Wrap(err), flow.debug), nil
	}

	// Retrieve the flow model from the database using the flow ID.
	flowModel, err := db.GetFlow(actionParam.FlowID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newFlowResultFromError(flow.errorState, ErrorOperationNotPermitted.Wrap(err), flow.debug), nil
		}
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	// Check if the flow has expired.
	if time.Now().After(flowModel.ExpiresAt) {
		return newFlowResultFromError(flow.errorState, ErrorFlowExpired, flow.debug), nil
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

	detail, err := flow.getStateDetail(flowModel.CurrentState)
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
	actionName := ActionName(actionParam.ActionName)

	// Get the action associated with the actionParam name.
	action, err := detail.actions.getByName(actionName)
	if err != nil {
		return newFlowResultFromError(flow.errorState, ErrorOperationNotPermitted.Wrap(err), flow.debug), nil
	}

	// Initialize the schema and action context for action execution.
	schema := newSchemaWithInputData(inputJSON)
	aic := defaultActionInitializationContext{schema: schema.toInitializationSchema(), stash: stash}
	action.Initialize(&aic)

	// Check if the action is suspended.
	if aic.isSuspended {
		return newFlowResultFromError(flow.errorState, ErrorOperationNotPermitted, flow.debug), nil
	}

	// Create a actionExecutionContext instance for action execution.
	aec := defaultActionExecutionContext{
		actionName:         actionName,
		input:              schema,
		defaultFlowContext: fc,
	}

	// Execute the action and handle any errors.
	err = action.Execute(&aec)
	if err != nil {
		return nil, fmt.Errorf("the action failed to handle the request: %w", err)
	}

	// Ensure that the action has set a result object.
	if aec.executionResult == nil {
		return nil, errors.New("the action has not set a result object")
	}

	// Generate a response based on the execution result.
	er := *aec.executionResult

	return er.generateResponse(fc, flow.debug), nil
}