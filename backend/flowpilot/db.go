package flowpilot

import (
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

// FlowModel represents the database model for a flow.
type FlowModel struct {
	ID        uuid.UUID // Unique ID of the flow.
	Data      string    // Stash data associated with the flow.
	CSRFToken string    // Current CSRF token
	Version   int       // Version of the flow.
	ExpiresAt time.Time // Expiry time of the flow.
	CreatedAt time.Time // Creation time of the flow.
	UpdatedAt time.Time // Update time of the flow.
}

// FlowDB is the interface for interacting with the flow database.
type FlowDB interface {
	GetFlow(flowID uuid.UUID) (*FlowModel, error)
	CreateFlow(flowModel FlowModel) error
	UpdateFlow(flowModel FlowModel) error
}

// flowDBWrapper is an extended FlowDB interface that includes additional methods.
type flowDBWrapper interface {
	FlowDB
	createFlowWithParam(p flowCreationParam) (*FlowModel, error)
	updateFlowWithParam(p flowUpdateParam) (*FlowModel, error)
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
	data      string    //
	csrfToken string    // Current CSRF token
	expiresAt time.Time // Expiry time of the flow.
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
		ID:        flowID,
		Data:      p.data,
		Version:   0,
		CSRFToken: p.csrfToken,
		ExpiresAt: p.expiresAt,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
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
	flowID    uuid.UUID // ID of the flow to update.
	data      string    // Updated stash data for the flow.
	version   int       // Updated version of the flow.
	csrfToken string    // Current CSRF tokens
	expiresAt time.Time // Updated expiry time of the flow.
	createdAt time.Time // Original creation time of the flow.
}

// UpdateFlowWithParam updates the specified flow with the given parameters.
func (w *defaultFlowDBWrapper) updateFlowWithParam(p flowUpdateParam) (*FlowModel, error) {
	// Prepare the updated FlowModel.
	fm := FlowModel{
		ID:        p.flowID,
		Data:      p.data,
		Version:   p.version,
		CSRFToken: p.csrfToken,
		ExpiresAt: p.expiresAt,
		UpdatedAt: time.Now().UTC(),
		CreatedAt: p.createdAt,
	}

	// Update the flow in the database.
	err := w.UpdateFlow(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to store updated flow to the dbw: %w", err)
	}

	return &fm, nil
}
