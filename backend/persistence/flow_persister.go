package persistence

import (
	"errors"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type FlowPersister interface {
	flowpilot.FlowDB
	Cleanup[models.Flow]
}

type flowPersister struct {
	tx *pop.Connection
}

func NewFlowPersister(tx *pop.Connection) FlowPersister {
	return flowPersister{tx}
}

func (p flowPersister) GetFlow(flowID uuid.UUID) (*flowpilot.FlowModel, error) {
	flowModel := models.Flow{}

	err := p.tx.Find(&flowModel, flowID)
	if err != nil {
		return nil, err
	}

	return flowModel.ToFlowpilotModel(), nil
}

func (p flowPersister) CreateFlow(flowModel flowpilot.FlowModel) error {
	f := models.Flow{
		ID:        flowModel.ID,
		Data:      flowModel.Data,
		Version:   flowModel.Version,
		CSRFToken: flowModel.CSRFToken,
		ExpiresAt: flowModel.ExpiresAt,
		CreatedAt: flowModel.CreatedAt,
		UpdatedAt: flowModel.UpdatedAt,
	}

	err := p.tx.Create(&f)
	if err != nil {
		return err
	}

	return nil
}

func (p flowPersister) UpdateFlow(flowModel flowpilot.FlowModel) error {
	f := &models.Flow{
		ID:        flowModel.ID,
		Data:      flowModel.Data,
		Version:   flowModel.Version,
		CSRFToken: flowModel.CSRFToken,
		ExpiresAt: flowModel.ExpiresAt,
		CreatedAt: flowModel.CreatedAt,
		UpdatedAt: flowModel.UpdatedAt,
	}

	previousVersion := flowModel.Version - 1

	count, err := p.tx.
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

func (p flowPersister) FindExpired(cutoffTime time.Time, page, perPage int) ([]models.Flow, error) {
	var items []models.Flow

	query := p.tx.
		Where("expires_at < ?", cutoffTime).
		Select("id").
		Paginate(page, perPage)
	err := query.All(&items)

	return items, err
}

func (p flowPersister) Delete(item models.Flow) error {
	return p.tx.Destroy(&item)
}
