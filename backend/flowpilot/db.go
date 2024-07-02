package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

// FlowModel represents the database model for a flow.
type FlowModel struct {
	ID            uuid.UUID  // Unique ID of the flow.
	CurrentState  StateName  // Current state of the flow.
	PreviousState *StateName // Previous state of the flow.
	StashData     string     // Stash data associated with the flow.
	CSRFToken     string     // Current CSRF token
	Version       int        // Version of the flow.
	ExpiresAt     time.Time  // Expiry time of the flow.
	CreatedAt     time.Time  // Creation time of the flow.
	UpdatedAt     time.Time  // Update time of the flow.
}

// TransitionModel represents the database model for a Transition.
type TransitionModel struct {
	ID        uuid.UUID  // Unique ID of the Transition.
	FlowID    uuid.UUID  // ID of the associated flow.
	Action    ActionName // Name of the action associated with the Action.
	FromState StateName  // Source state of the Transition.
	ToState   StateName  // Target state of the Transition.
	InputData string     // Input data associated with the Transition.
	ErrorCode *string    // Optional error code associated with the Transition.
	CreatedAt time.Time  // Creation time of the Transition.
	UpdatedAt time.Time  // Update time of the Transition.
}

// FlowDB is the interface for interacting with the flow database.
type FlowDB interface {
	GetFlow(flowID uuid.UUID) (*FlowModel, error)
	CreateFlow(flowModel FlowModel) error
	UpdateFlow(flowModel FlowModel) error
	CreateTransition(transitionModel TransitionModel) error

	// TODO: "FindLastTransitionWithAction" might be useless, or can be replaced.

	FindLastTransitionWithAction(flowID uuid.UUID, method ActionName) (*TransitionModel, error)
}

// flowDBWrapper is an extended FlowDB interface that includes additional methods.
type flowDBWrapper interface {
	FlowDB
	createFlowWithParam(p flowCreationParam) (*FlowModel, error)
	updateFlowWithParam(p flowUpdateParam) (*FlowModel, error)
	createTransitionWithParam(p transitionCreationParam) (*TransitionModel, error)
}

// defaultFlowDBWrapper wraps a FlowDB instance to provide additional functionality.
type defaultFlowDBWrapper struct {
	FlowDB
}

// wrapDB wraps a FlowDB instance to provide flowDBWrapper functionality.
func wrapDB(db FlowDB) flowDBWrapper {
	return &defaultFlowDBWrapper{FlowDB: db}
}

// flowCreationParam holds parameters for creating a new flow.
type flowCreationParam struct {
	currentState StateName // Initial state  of the flow.
	stash        string    //
	csrfToken    string    // Current CSRF token
	expiresAt    time.Time // Expiry time of the flow.
}

// CreateFlowWithParam creates a new flow with the given parameters.
func (w *defaultFlowDBWrapper) createFlowWithParam(p flowCreationParam) (*FlowModel, error) {
	// Generate a new UUID for the flow.
	flowID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate a new flow id: %w", err)
	}

	// Prepare the FlowModel for creation.
	fm := FlowModel{
		ID:            flowID,
		CurrentState:  p.currentState,
		PreviousState: nil,
		StashData:     p.stash,
		Version:       0,
		CSRFToken:     p.csrfToken,
		ExpiresAt:     p.expiresAt,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create the flow in the database.
	err = w.CreateFlow(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to store a new flow to the dbw: %w", err)
	}

	return &fm, nil
}

// flowUpdateParam holds parameters for updating a flow.
type flowUpdateParam struct {
	flowID        uuid.UUID // ID of the flow to update.
	nextState     StateName // Next state of the flow.
	previousState StateName // Previous state of the flow.
	stashData     string    // Updated stash data for the flow.
	version       int       // Updated version of the flow.
	csrfToken     string    // Current CSRF tokens
	expiresAt     time.Time // Updated expiry time of the flow.
	createdAt     time.Time // Original creation time of the flow.
}

// UpdateFlowWithParam updates the specified flow with the given parameters.
func (w *defaultFlowDBWrapper) updateFlowWithParam(p flowUpdateParam) (*FlowModel, error) {
	// Prepare the updated FlowModel.
	fm := FlowModel{
		ID:            p.flowID,
		CurrentState:  p.nextState,
		PreviousState: &p.previousState,
		StashData:     p.stashData,
		Version:       p.version,
		CSRFToken:     p.csrfToken,
		ExpiresAt:     p.expiresAt,
		UpdatedAt:     time.Now().UTC(),
		CreatedAt:     p.createdAt,
	}

	// Update the flow in the database.
	err := w.UpdateFlow(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to store updated flow to the dbw: %w", err)
	}

	return &fm, nil
}

// transitionCreationParam holds parameters for creating a new Transition.
type transitionCreationParam struct {
	flowID     uuid.UUID  // ID of the associated flow.
	actionName ActionName // Name of the action associated with the Transition.
	fromState  StateName  // Source state of the Transition.
	toState    StateName  // Target state of the Transition.
	inputData  string     // Input data associated with the Transition.
	flowError  FlowError  // Optional flowError associated with the Transition.
}

// CreateTransitionWithParam creates a new Transition with the given parameters.
func (w *defaultFlowDBWrapper) createTransitionWithParam(p transitionCreationParam) (*TransitionModel, error) {
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
