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

// methodInitializationContext represents the basic context for a Flow method's initialization.
type methodInitializationContext interface {
	// AddInputs adds input parameters to the schema.
	AddInputs(inputList ...Input)
	// Stash returns the ReadOnlyJSONManager for accessing stash data.
	Stash() jsonmanager.ReadOnlyJSONManager
	// SuspendMethod suspends the current method's execution.
	SuspendMethod()
}

// methodExecutionContext represents the context for a method execution.
type methodExecutionContext interface {
	flowContext
	// Input returns the MethodExecutionSchema for the method.
	Input() MethodExecutionSchema

	// TODO: FetchMethodInput (for a method name) is maybe useless and can be removed or replaced.

	// FetchMethodInput fetches input data for a specific method.
	FetchMethodInput(methodName MethodName) (jsonmanager.ReadOnlyJSONManager, error)
	// ValidateInputData validates the input data against the schema.
	ValidateInputData() bool
	// CopyInputValuesToStash copies specified inputs to the stash.
	CopyInputValuesToStash(inputNames ...string) error
}

// methodExecutionContinuationContext represents the context within a method continuation.
type methodExecutionContinuationContext interface {
	methodExecutionContext
	// ContinueFlow continues the Flow execution to the specified next state.
	ContinueFlow(nextState StateName) error
	// ContinueFlowWithError continues the Flow execution to the specified next state with an error.
	ContinueFlowWithError(nextState StateName, flowErr FlowError) error

	// TODO: Implement a function to step back to the previous state (while skipping self-transitions and recalling preserved data).
}

// InitializationContext is a shorthand for methodInitializationContext within flow methods.
type InitializationContext interface {
	methodInitializationContext
}

// ExecutionContext is a shorthand for methodExecutionContinuationContext within flow methods.
type ExecutionContext interface {
	methodExecutionContinuationContext
}

// TODO: The following interfaces are meant for a plugin system. #tbd

// PluginBeforeMethodExecutionContext represents the context for a plugin before a method execution.
type PluginBeforeMethodExecutionContext interface {
	methodExecutionContinuationContext
}

// PluginAfterMethodExecutionContext represents the context for a plugin after a method execution.
type PluginAfterMethodExecutionContext interface {
	methodExecutionContext
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

// executeFlowMethod processes the Flow and returns a Response.
func executeFlowMethod(db FlowDB, flow Flow, options flowExecutionOptions) (FlowResult, error) {
	// Parse the action parameter to get the method name and Flow ID.
	action, err := utils.ParseActionParam(options.action)
	if err != nil {
		return newFlowResultFromError(flow.ErrorState, ErrorActionParamInvalid.Wrap(err), flow.Debug), nil
	}

	// Retrieve the Flow model from the database using the Flow ID.
	flowModel, err := db.GetFlow(action.FlowID)
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

	// Create a MethodName from the parsed action method name.
	methodName := MethodName(action.MethodName)
	// Get the method associated with the action method name.
	method, err := transitions.getMethod(methodName)
	if err != nil {
		return newFlowResultFromError(flow.ErrorState, ErrorOperationNotPermitted.Wrap(err), flow.Debug), nil
	}

	// Initialize the schema and method context for method execution.
	schema := newSchemaWithInputData(inputJSON)
	mic := defaultMethodInitializationContext{schema: schema.toInitializationSchema(), stash: stash}
	method.Initialize(&mic)

	// Check if the method is suspended.
	if mic.isSuspended {
		return newFlowResultFromError(flow.ErrorState, ErrorOperationNotPermitted, flow.Debug), nil
	}

	// Create a methodExecutionContext instance for method execution.
	mec := defaultMethodExecutionContext{
		methodName:         methodName,
		input:              schema,
		defaultFlowContext: fc,
	}

	// Execute the method and handle any errors.
	err = method.Execute(&mec)
	if err != nil {
		return nil, fmt.Errorf("the method failed to handle the request: %w", err)
	}

	// Ensure that the method has set a result object.
	if mec.methodResult == nil {
		return nil, errors.New("the method has not set a result object")
	}

	// Generate a response based on the execution result.
	er := *mec.methodResult

	return er.generateResponse(fc), nil
}
