package models

import (
	"errors"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type FlowDB struct {
	tx *pop.Connection
}

func NewFlowDB(tx *pop.Connection) flowpilot.FlowDB {
	return FlowDB{tx}
}

func (flowDB FlowDB) GetFlow(flowID uuid.UUID) (*flowpilot.FlowModel, error) {
	flowModel := Flow{}

	err := flowDB.tx.Find(&flowModel, flowID)
	if err != nil {
		return nil, err
	}

	return flowModel.ToFlowpilotModel(), nil
}

func (flowDB FlowDB) CreateFlow(flowModel flowpilot.FlowModel) error {
	f := Flow{
		ID:        flowModel.ID,
		Data:      flowModel.Data,
		Version:   flowModel.Version,
		CSRFToken: flowModel.CSRFToken,
		ExpiresAt: flowModel.ExpiresAt,
		CreatedAt: flowModel.CreatedAt,
		UpdatedAt: flowModel.UpdatedAt,
	}

	err := flowDB.tx.Create(&f)
	if err != nil {
		return err
	}

	return nil
}

func (flowDB FlowDB) UpdateFlow(flowModel flowpilot.FlowModel) error {
	f := &Flow{
		ID:        flowModel.ID,
		Data:      flowModel.Data,
		Version:   flowModel.Version,
		CSRFToken: flowModel.CSRFToken,
		ExpiresAt: flowModel.ExpiresAt,
		CreatedAt: flowModel.CreatedAt,
		UpdatedAt: flowModel.UpdatedAt,
	}

	previousVersion := flowModel.Version - 1

	count, err := flowDB.tx.
		Where("id = ?", f.ID).
		Where("version = ?", previousVersion).
		UpdateQuery(f, "version", "csrf_token", "data")
	if err != nil {
		return err
	}

	if count != 1 {
		return errors.New("version conflict while updating the flow")
	}

	return nil
}
