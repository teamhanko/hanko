package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

// FlowModel represents the database model for a defaultFlow.
type FlowModel struct {
	ID           uuid.UUID // Unique ID of the defaultFlow.
	CurrentState StateName // Current addState of the defaultFlow.
	StashData    string    // Stash data associated with the defaultFlow.
	Version      int       // Version of the defaultFlow.
	Completed    bool      // Flag indicating if the defaultFlow is completed.
	ExpiresAt    time.Time // Expiry time of the defaultFlow.
	CreatedAt    time.Time // Creation time of the defaultFlow.
	UpdatedAt    time.Time // Update time of the defaultFlow.
}

// TransitionModel represents the database model for a Transition.
type TransitionModel struct {
	ID        uuid.UUID  // Unique ID of the Transition.
	FlowID    uuid.UUID  // ID of the associated defaultFlow.
	Action    ActionName // Name of the action associated with the Action.
	FromState StateName  // Source addState of the Transition.
	ToState   StateName  // Target addState of the Transition.
	InputData string     // Input data associated with the Transition.
	ErrorCode *string    // Optional error code associated with the Transition.
	CreatedAt time.Time  // Creation time of the Transition.
	UpdatedAt time.Time  // Update time of the Transition.
}

// FlowDB is the interface for interacting with the defaultFlow database.
type FlowDB interface {
	GetFlow(flowID uuid.UUID) (*FlowModel, error)
	CreateFlow(flowModel FlowModel) error
	UpdateFlow(flowModel FlowModel) error
	CreateTransition(transitionModel TransitionModel) error

	// TODO: "FindLastTransitionWithAction" might be useless, or can be replaced.

	FindLastTransitionWithAction(flowID uuid.UUID, method ActionName) (*TransitionModel, error)
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

// flowCreationParam holds parameters for creating a new defaultFlow.
type flowCreationParam struct {
	currentState StateName // Initial addState of the defaultFlow.
	expiresAt    time.Time // Expiry time of the defaultFlow.
}

// CreateFlowWithParam creates a new defaultFlow with the given parameters.
func (w *DefaultFlowDBWrapper) CreateFlowWithParam(p flowCreationParam) (*FlowModel, error) {
	// Generate a new UUID for the defaultFlow.
	flowID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate a new defaultFlow id: %w", err)
	}

	// Prepare the FlowModel for creation.
	fm := FlowModel{
		ID:           flowID,
		CurrentState: p.currentState,
		StashData:    "{}",
		Version:      0,
		Completed:    false,
		ExpiresAt:    p.expiresAt,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// Create the defaultFlow in the database.
	err = w.CreateFlow(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to store a new defaultFlow to the dbw: %w", err)
	}

	return &fm, nil
}

// flowUpdateParam holds parameters for updating a defaultFlow.
type flowUpdateParam struct {
	flowID    uuid.UUID // ID of the flow to update.
	nextState StateName // Next addState of the flow.
	stashData string    // Updated stash data for the flow.
	version   int       // Updated version of the flow.
	completed bool      // Flag indicating if the flow is completed.
	expiresAt time.Time // Updated expiry time of the flow.
	createdAt time.Time // Original creation time of the flow.
}

// UpdateFlowWithParam updates the specified defaultFlow with the given parameters.
func (w *DefaultFlowDBWrapper) UpdateFlowWithParam(p flowUpdateParam) (*FlowModel, error) {
	// Prepare the updated FlowModel.
	fm := FlowModel{
		ID:           p.flowID,
		CurrentState: p.nextState,
		StashData:    p.stashData,
		Version:      p.version,
		Completed:    p.completed,
		ExpiresAt:    p.expiresAt,
		UpdatedAt:    time.Now().UTC(),
		CreatedAt:    p.createdAt,
	}

	// Update the defaultFlow in the database.
	err := w.UpdateFlow(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to store updated flow to the dbw: %w", err)
	}

	return &fm, nil
}

// transitionCreationParam holds parameters for creating a new Transition.
type transitionCreationParam struct {
	flowID     uuid.UUID  // ID of the associated defaultFlow.
	actionName ActionName // Name of the action associated with the Transition.
	fromState  StateName  // Source addState of the Transition.
	toState    StateName  // Target addState of the Transition.
	inputData  string     // Input data associated with the Transition.
	flowError  FlowError  // Optional flowError associated with the Transition.
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
		Action:    p.actionName,
		FromState: p.fromState,
		ToState:   p.toState,
		InputData: p.inputData,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Set the error code if provided.
	if p.flowError != nil {
		code := p.flowError.Code()
		tm.ErrorCode = &code
	}

	// Create the Transition in the database.
	err = w.CreateTransition(tm)
	if err != nil {
		return nil, fmt.Errorf("failed to store a new Transition to the dbw: %w", err)
	}

	return &tm, err
}
