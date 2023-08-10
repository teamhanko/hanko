package flowpilot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"hanko_flowsc/flowpilot/jsonmanager"
	"hanko_flowsc/flowpilot/utils"
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
	// Flash returns the JSONManager for accessing flash data.
	Flash() jsonmanager.JSONManager
	// GetInitialState returns the initial state of the Flow.
	GetInitialState() StateName
	// GetCurrentState returns the current state of the Flow.
	GetCurrentState() StateName
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
	AddInputs(inputList ...*DefaultInput) *DefaultSchema
	// Stash returns the ReadOnlyJSONManager for accessing stash data.
	Stash() jsonmanager.ReadOnlyJSONManager
	// SuspendMethod suspends the current method's execution.
	SuspendMethod()
}

// methodExecutionContext represents the context for a method execution.
type methodExecutionContext interface {
	// Input returns the ReadOnlyJSONManager for accessing input data.
	Input() jsonmanager.ReadOnlyJSONManager
	// Schema returns the MethodExecutionSchema for the method.
	Schema() MethodExecutionSchema

	// TODO: FetchMethodInput (for a method name) is maybe useless and can be removed or replaced.

	// FetchMethodInput fetches input data for a specific method.
	FetchMethodInput(methodName MethodName) (jsonmanager.ReadOnlyJSONManager, error)
	// ValidateInputData validates the input data against the schema.
	ValidateInputData() bool

	// TODO: CopyInputsToStash can maybe removed or replaced with an option you can set via the input options.

	// CopyInputsToStash copies specified inputs to the stash.
	CopyInputsToStash(inputNames ...string) error
}

// methodExecutionContinuationContext represents the context within a method continuation.
type methodExecutionContinuationContext interface {
	flowContext
	methodExecutionContext
	// ContinueFlow continues the Flow execution to the specified next state.
	ContinueFlow(nextState StateName) error
	// ContinueFlowWithError continues the Flow execution to the specified next state with an error.
	ContinueFlowWithError(nextState StateName, errType *ErrorType) error

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
func createAndInitializeFlow(db FlowDB, flow Flow) (*Response, error) {
	// Wrap the provided FlowDB with additional functionality.
	dbw := wrapDB(db)
	// Calculate the expiration time for the Flow.
	expiresAt := time.Now().Add(flow.TTL).UTC()

	// Initialize JSONManagers for stash, payload, and flash data.

	// TODO: Consider implementing types for stash, payload, and flash that extend "jsonmanager.NewJSONManager()".
	// This could enhance the code structure and provide clearer interfaces for handling these data structures.

	stash := jsonmanager.NewJSONManager()
	payload := jsonmanager.NewJSONManager()
	flash := jsonmanager.NewJSONManager()

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
		flash:     flash,
	}

	// Generate a response based on the execution result.
	er := executionResult{nextState: flowModel.CurrentState}
	return er.generateResponse(fc)
}

// executeFlowMethod processes the Flow and returns a Response.
func executeFlowMethod(db FlowDB, flow Flow, options flowExecutionOptions) (*Response, error) {
	// Parse the action parameter to get the method name and Flow ID.
	action, err := utils.ParseActionParam(options.action)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action param: %w", err)
	}

	// Retrieve the Flow model from the database using the Flow ID.
	flowModel, err := db.GetFlow(action.FlowID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &Response{State: flow.ErrorState, Error: OperationNotPermittedError}, nil
		}
		return nil, fmt.Errorf("failed to get Flow: %w", err)
	}

	// Check if the Flow has expired.
	if time.Now().After(flowModel.ExpiresAt) {
		return &Response{State: flow.ErrorState, Error: FlowExpiredError}, nil
	}

	// Parse stash data from the Flow model.
	stash, err := jsonmanager.NewJSONManagerFromString(flowModel.StashData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stash from flowModel: %w", err)
	}

	// Initialize JSONManagers for payload and flash data.
	payload := jsonmanager.NewJSONManager()
	flash := jsonmanager.NewJSONManager()

	// Create a defaultFlowContext instance.
	fc := defaultFlowContext{
		flow:      flow,
		dbw:       wrapDB(db),
		flowModel: *flowModel,
		stash:     stash,
		payload:   payload,
		flash:     flash,
	}

	// Get the available transitions for the current state.
	transitions := fc.getCurrentTransitions()
	if transitions == nil {
		return &Response{State: flow.ErrorState, Error: OperationNotPermittedError}, nil
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
		return &Response{State: flow.ErrorState, Error: OperationNotPermittedError}, nil
	}

	// Initialize the schema and method context for method execution.
	schema := DefaultSchema{}
	mic := defaultMethodInitializationContext{schema: &schema, stash: stash}
	method.Initialize(&mic)

	// Check if the method is suspended.
	if mic.isSuspended {
		return &Response{State: flow.ErrorState, Error: OperationNotPermittedError}, nil
	}

	// Create a methodExecutionContext instance for method execution.
	mec := defaultMethodExecutionContext{
		schema:             &schema,
		methodName:         methodName,
		input:              inputJSON,
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
	return er.generateResponse(fc)
}
