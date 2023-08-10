package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

// FlowModel represents the database model for a Flow.
type FlowModel struct {
	ID            uuid.UUID  // Unique ID of the Flow.
	CurrentState  StateName  // Current state of the Flow.
	PreviousState *StateName // Previous state of the Flow.
	StashData     string     // Stash data associated with the Flow.
	Version       int        // Version of the Flow.
	Completed     bool       // Flag indicating if the Flow is completed.
	ExpiresAt     time.Time  // Expiry time of the Flow.
	CreatedAt     time.Time  // Creation time of the Flow.
	UpdatedAt     time.Time  // Update time of the Flow.
}

// TransitionModel represents the database model for a Transition.
type TransitionModel struct {
	ID        uuid.UUID  // Unique ID of the Transition.
	FlowID    uuid.UUID  // ID of the associated Flow.
	Method    MethodName // Name of the method associated with the Transition.
	FromState StateName  // Source state of the Transition.
	ToState   StateName  // Target state of the Transition.
	InputData string     // Input data associated with the Transition.
	ErrorCode *string    // Optional error code associated with the Transition.
	CreatedAt time.Time  // Creation time of the Transition.
	UpdatedAt time.Time  // Update time of the Transition.
}

// FlowDB is the interface for interacting with the Flow database.
type FlowDB interface {
	GetFlow(flowID uuid.UUID) (*FlowModel, error)
	CreateFlow(flowModel FlowModel) error
	UpdateFlow(flowModel FlowModel) error
	CreateTransition(transitionModel TransitionModel) error

	// TODO: "FindLastTransitionWithMethod" might be useless, or can be replaced.

	FindLastTransitionWithMethod(flowID uuid.UUID, method MethodName) (*TransitionModel, error)
}

// FlowDBWrapper is an extended FlowDB interface that includes additional methods.
type FlowDBWrapper interface {
	FlowDB
	CreateFlowWithParam(p flowCreationParam) (*FlowModel, error)
	UpdateFlowWithParam(p flowUpdateParam) (*FlowModel, error)
	CreateTransitionWithParam(p transitionCreationParam) (*TransitionModel, error)
}

// DefaultFlowDBWrapper wraps a FlowDB instance to provide additional functionality.
type DefaultFlowDBWrapper struct {
	FlowDB
}

// wrapDB wraps a FlowDB instance to provide FlowDBWrapper functionality.
func wrapDB(db FlowDB) FlowDBWrapper {
	return &DefaultFlowDBWrapper{FlowDB: db}
}

// flowCreationParam holds parameters for creating a new Flow.
type flowCreationParam struct {
	currentState StateName // Initial state of the Flow.
	expiresAt    time.Time // Expiry time of the Flow.
}

// CreateFlowWithParam creates a new Flow with the given parameters.
func (w *DefaultFlowDBWrapper) CreateFlowWithParam(p flowCreationParam) (*FlowModel, error) {
	// Generate a new UUID for the Flow.
	flowID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate a new Flow id: %w", err)
	}

	// Prepare the FlowModel for creation.
	fm := FlowModel{
		ID:            flowID,
		CurrentState:  p.currentState,
		PreviousState: nil,
		StashData:     "{}",
		Version:       0,
		Completed:     false,
		ExpiresAt:     p.expiresAt,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create the Flow in the database.
	err = w.CreateFlow(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to store a new Flow to the dbw: %w", err)
	}

	return &fm, nil
}

// flowUpdateParam holds parameters for updating a Flow.
type flowUpdateParam struct {
	flowID        uuid.UUID  // ID of the Flow to update.
	nextState     StateName  // Next state of the Flow.
	previousState *StateName // Previous state of the Flow.
	stashData     string     // Updated stash data for the Flow.
	version       int        // Updated version of the Flow.
	completed     bool       // Flag indicating if the Flow is completed.
	expiresAt     time.Time  // Updated expiry time of the Flow.
	createdAt     time.Time  // Original creation time of the Flow.
}

// UpdateFlowWithParam updates the specified Flow with the given parameters.
func (w *DefaultFlowDBWrapper) UpdateFlowWithParam(p flowUpdateParam) (*FlowModel, error) {
	// Prepare the updated FlowModel.
	fm := FlowModel{
		ID:            p.flowID,
		CurrentState:  p.nextState,
		PreviousState: p.previousState,
		StashData:     p.stashData,
		Version:       p.version,
		Completed:     p.completed,
		ExpiresAt:     p.expiresAt,
		UpdatedAt:     time.Now().UTC(),
		CreatedAt:     p.createdAt,
	}

	// Update the Flow in the database.
	err := w.UpdateFlow(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to store updated Flow to the dbw: %w", err)
	}

	return &fm, nil
}

// transitionCreationParam holds parameters for creating a new Transition.
type transitionCreationParam struct {
	flowID     uuid.UUID  // ID of the associated Flow.
	methodName MethodName // Name of the method associated with the Transition.
	fromState  StateName  // Source state of the Transition.
	toState    StateName  // Target state of the Transition.
	inputData  string     // Input data associated with the Transition.
	errType    *ErrorType // Optional error type associated with the Transition.
}

// CreateTransitionWithParam creates a new Transition with the given parameters.
func (w *DefaultFlowDBWrapper) CreateTransitionWithParam(p transitionCreationParam) (*TransitionModel, error) {
	// Generate a new UUID for the Transition.
	transitionID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new Transition id: %w", err)
	}

	// Prepare the TransitionModel for creation.
	tm := TransitionModel{
		ID:        transitionID,
		FlowID:    p.flowID,
		Method:    p.methodName,
		FromState: p.fromState,
		ToState:   p.toState,
		InputData: p.inputData,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Set the error code if provided.
	if p.errType != nil {
		tm.ErrorCode = &p.errType.Code
	}

	// Create the Transition in the database.
	err = w.CreateTransition(tm)
	if err != nil {
		return nil, fmt.Errorf("failed to store a new Transition to the dbw: %w", err)
	}

	return &tm, err
}
