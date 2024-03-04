package models

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Flow is used by pop to map your flows database table to your go code.
type Flow struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CurrentState string    `json:"current_state" db:"current_state"`
	StashData    string    `json:"stash_data" db:"stash_data"`
	Version      int       `json:"version" db:"version"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	// transitions  transitions `json:"transitions" has_many:"transitions" order_by:"created_at desc"`
}

func (f *Flow) ToFlowpilotModel() *flowpilot.FlowModel {
	flow := flowpilot.FlowModel{
		ID:           f.ID,
		CurrentState: flowpilot.StateName(f.CurrentState),
		StashData:    f.StashData,
		Version:      f.Version,
		ExpiresAt:    f.ExpiresAt,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}

	return &flow
}

// Flows is not required by pop and may be deleted
type Flows []Flow

// Validate gets run every time you call a "pop.validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (f *Flow) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (f *Flow) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (f *Flow) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
