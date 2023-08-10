package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"time"
)

// Transition is used by pop to map your Transitions database table to your go code.
type Transition struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FlowID    uuid.UUID `json:"-" db:"flow_id" `
	Method    string    `json:"method" db:"method"`
	FromState string    `json:"from_state" db:"from_state"`
	ToState   string    `json:"to_state" db:"to_state"`
	InputData string    `json:"input_data" db:"input_data"`
	ErrorCode *string   `json:"error_code" db:"error_code"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	// flow      flow      `json:"flow,omitempty" belongs_to:"flow"`
}

func (t *Transition) ToFlowpilotModel() *flowpilot.TransitionModel {
	return &flowpilot.TransitionModel{
		ID:        t.ID,
		FlowID:    t.FlowID,
		Method:    flowpilot.MethodName(t.Method),
		FromState: flowpilot.StateName(t.FromState),
		ToState:   flowpilot.StateName(t.ToState),
		InputData: t.InputData,
		ErrorCode: t.ErrorCode,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// Transitions is not required by pop and may be deleted
type Transitions []Transition

// Validate gets run every time you call a "pop.validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Transition) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Transition) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Transition) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
