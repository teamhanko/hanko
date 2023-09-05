package flowpilot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
	"time"
)

// flowContext represents the basic context for a Flow.
type flowContext interface {
	// GetFlowID returns the unique ID of the current Flow.
	GetFlowID() uuid.UUID
	// GetPath returns the current path within the Flow.
	GetPath() string
	// Payload returns the JSONManager for accessing payload data.
	Payload() jsonmanager.JSONManager
	// Stash returns the JSONManager for accessing stash data.
	Stash() jsonmanager.JSONManager
	// GetInitialState returns the initial state of the Flow.
	GetInitialState() StateName
	// GetCurrentState returns the current state of the Flow.
	GetCurrentState() StateName
	// CurrentStateEquals returns true, when one of the given states matches the current state.
	CurrentStateEquals(states ...StateName) bool
	// GetPreviousState returns the previous state of the Flow.
	GetPreviousState() *StateName
	// GetErrorState returns the designated error state of the Flow.
	GetErrorState() StateName
	// GetEndState returns the final state of the Flow.
	GetEndState() StateName
	// StateExists checks if a given state exists within the Flow.
	StateExists(stateName StateName) bool
}

// actionInitializationContext represents the basic context for a Flow action's initialization.
type actionInitializationContext interface {
	// AddInputs adds input parameters to the schema.
	AddInputs(inputList ...Input)
	// Stash returns the ReadOnlyJSONManager for accessing stash data.
	Stash() jsonmanager.ReadOnlyJSONManager
	// SuspendAction suspends the current action's execution.
	SuspendAction()
}

// actionExecutionContext represents the context for a action execution.
type actionExecutionContext interface {
	flowContext
	// Input returns the ExecutionSchema for the action.
	Input() ExecutionSchema

	// TODO: FetchActionInput (for a action name) is maybe useless and can be removed or replaced.

	// FetchActionInput fetches input data for a specific action.
	FetchActionInput(actionName ActionName) (jsonmanager.ReadOnlyJSONManager, error)
	// ValidateInputData validates the input data against the schema.
	ValidateInputData() bool
	// CopyInputValuesToStash copies specified inputs to the stash.
	CopyInputValuesToStash(inputNames ...string) error
}

// actionExecutionContinuationContext represents the context within an action continuation.
type actionExecutionContinuationContext interface {
	actionExecutionContext
	// ContinueFlow continues the Flow execution to the specified next state.
	ContinueFlow(nextState StateName) error
	// ContinueFlowWithError continues the Flow execution to the specified next state with an error.
	ContinueFlowWithError(nextState StateName, flowErr FlowError) error

	// TODO: Implement a function to step back to the previous state (while skipping self-transitions and recalling preserved data).
}

// InitializationContext is a shorthand for actionInitializationContext within flow actions.
type InitializationContext interface {
	actionInitializationContext
}

// ExecutionContext is a shorthand for actionExecutionContinuationContext within flow actions.
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

// createAndInitializeFlow initializes the Flow and returns a flow Response.
func createAndInitializeFlow(db FlowDB, flow Flow) (FlowResult, error) {
	// Wrap the provided FlowDB with additional functionality.
	dbw := wrapDB(db)
	// Calculate the expiration time for the Flow.
	expiresAt := time.Now().Add(flow.TTL).UTC()

	// TODO: Consider implementing types for stash and payload that extend "jsonmanager.NewJSONManager()".
	// This could enhance the code structure and provide clearer interfaces for handling these data structures.

	// Initialize JSONManagers for stash and payload.
	stash := jsonmanager.NewJSONManager()
	payload := jsonmanager.NewJSONManager()

	// Create a new Flow model with the provided parameters.
	p := flowCreationParam{currentState: flow.InitialState, expiresAt: expiresAt}
	flowModel, err := dbw.CreateFlowWithParam(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create Flow: %w", err)
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

	return er.generateResponse(fc), nil
}

// executeFlowAction processes the Flow and returns a Response.
func executeFlowAction(db FlowDB, flow Flow, options flowExecutionOptions) (FlowResult, error) {
	// Parse the actionParam parameter to get the actionParam name and Flow ID.
	actionParam, err := utils.ParseActionParam(options.action)
	if err != nil {
		return newFlowResultFromError(flow.ErrorState, ErrorActionParamInvalid.Wrap(err), flow.Debug), nil
	}

	// Retrieve the Flow model from the database using the Flow ID.
	flowModel, err := db.GetFlow(actionParam.FlowID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newFlowResultFromError(flow.ErrorState, ErrorOperationNotPermitted.Wrap(err), flow.Debug), nil
		}
		return nil, fmt.Errorf("failed to get Flow: %w", err)
	}

	// Check if the Flow has expired.
	if time.Now().After(flowModel.ExpiresAt) {
		return newFlowResultFromError(flow.ErrorState, ErrorFlowExpired, flow.Debug), nil
	}

	// Parse stash data from the Flow model.
	stash, err := jsonmanager.NewJSONManagerFromString(flowModel.StashData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stash from flowModel: %w", err)
	}

	// Initialize JSONManagers for payload and flash data.
	payload := jsonmanager.NewJSONManager()

	// Create a defaultFlowContext instance.
	fc := defaultFlowContext{
		flow:      flow,
		dbw:       wrapDB(db),
		flowModel: *flowModel,
		stash:     stash,
		payload:   payload,
	}

	// Get the available transitions for the current state.
	transitions := fc.getCurrentTransitions()
	if transitions == nil {
		err2 := errors.New("the state does not allow to continue with the flow")
		return newFlowResultFromError(flow.ErrorState, ErrorOperationNotPermitted.Wrap(err2), flow.Debug), nil
	}

	// Parse raw input data into JSONManager.
	raw := options.inputData.getJSONStringOrDefault()
	inputJSON, err := jsonmanager.NewJSONManagerFromString(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input data: %w", err)
	}

	// Create a ActionName from the parsed actionParam name.
	actionName := ActionName(actionParam.ActionName)
	// Get the action associated with the actionParam name.
	action, err := transitions.getAction(actionName)
	if err != nil {
		return newFlowResultFromError(flow.ErrorState, ErrorOperationNotPermitted.Wrap(err), flow.Debug), nil
	}

	// Initialize the schema and action context for action execution.
	schema := newSchemaWithInputData(inputJSON)
	mic := defaultActionInitializationContext{schema: schema.toInitializationSchema(), stash: stash}
	action.Initialize(&mic)

	// Check if the action is suspended.
	if mic.isSuspended {
		return newFlowResultFromError(flow.ErrorState, ErrorOperationNotPermitted, flow.Debug), nil
	}

	// Create a actionExecutionContext instance for action execution.
	mec := defaultActionExecutionContext{
		actionName:         actionName,
		input:              schema,
		defaultFlowContext: fc,
	}

	// Execute the action and handle any errors.
	err = action.Execute(&mec)
	if err != nil {
		return nil, fmt.Errorf("the action failed to handle the request: %w", err)
	}

	// Ensure that the action has set a result object.
	if mec.executionResult == nil {
		return nil, errors.New("the action has not set a result object")
	}

	// Generate a response based on the execution result.
	er := *mec.executionResult

	return er.generateResponse(fc), nil
}
