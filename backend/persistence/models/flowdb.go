package models

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type FlowTestUser struct {
	ID                 string
	Email              string
	Username           string
	Password           string
	Passcode2faEnabled bool
	PasskeySynced      bool
}

type FlowTestUserList []FlowTestUser

func (ul FlowTestUserList) FindByID(id string) (*FlowTestUser, error) {
	for _, user := range ul {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user with ID %s not found", id)
}

func (ul FlowTestUserList) FindByEmail(email string) (*FlowTestUser, error) {
	for _, user := range ul {
		if user.Email == email {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user with email %s not found", email)
}

func (ul FlowTestUserList) FindByUsername(username string) (*FlowTestUser, error) {
	for _, user := range ul {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user with username %s not found", username)
}

var MyUsers = FlowTestUserList{
	{
		ID:                 "a1b229c0-a1e3-44de-b770-152e18abb31c",
		Email:              "user1@example.com",
		Username:           "user1",
		Password:           "test",
		Passcode2faEnabled: false,
		PasskeySynced:      false,
	},
	{
		ID:                 "26a70349-1136-4c3f-b7dd-2725e872b357",
		Email:              "user2@example.com",
		Username:           "user2",
		Password:           "test",
		Passcode2faEnabled: true,
		PasskeySynced:      true,
	},
}

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
		ID:            flowModel.ID,
		CurrentState:  string(flowModel.CurrentState),
		PreviousState: nil,
		StashData:     flowModel.StashData,
		Version:       flowModel.Version,
		Completed:     flowModel.Completed,
		ExpiresAt:     flowModel.ExpiresAt,
		CreatedAt:     flowModel.CreatedAt,
		UpdatedAt:     flowModel.UpdatedAt,
	}

	err := flowDB.tx.Create(&f)
	if err != nil {
		return err
	}

	return nil
}

func (flowDB FlowDB) UpdateFlow(flowModel flowpilot.FlowModel) error {
	f := &Flow{
		ID:           flowModel.ID,
		CurrentState: string(flowModel.CurrentState),
		StashData:    flowModel.StashData,
		Version:      flowModel.Version,
		Completed:    flowModel.Completed,
		ExpiresAt:    flowModel.ExpiresAt,
		CreatedAt:    flowModel.CreatedAt,
		UpdatedAt:    flowModel.UpdatedAt,
	}

	if ps := flowModel.PreviousState; ps != nil {
		previousState := string(*ps)
		f.PreviousState = &previousState
	}

	previousVersion := flowModel.Version - 1

	count, err := flowDB.tx.
		Where("id = ?", f.ID).
		Where("version = ?", previousVersion).
		UpdateQuery(f, "current_state", "previous_state", "stash_data", "version", "completed")
	if err != nil {
		return err
	}

	if count != 1 {
		return errors.New("version conflict while updating the flow")
	}

	return nil
}

func (flowDB FlowDB) CreateTransition(transitionModel flowpilot.TransitionModel) error {
	t := Transition{
		ID:        transitionModel.ID,
		FlowID:    transitionModel.FlowID,
		Action:    string(transitionModel.Action),
		FromState: string(transitionModel.FromState),
		ToState:   string(transitionModel.ToState),
		InputData: transitionModel.InputData,
		ErrorCode: transitionModel.ErrorCode,
		CreatedAt: transitionModel.CreatedAt,
		UpdatedAt: transitionModel.UpdatedAt,
	}

	err := flowDB.tx.Create(&t)
	if err != nil {
		return err
	}

	return nil
}

func (flowDB FlowDB) FindLastTransitionWithAction(flowID uuid.UUID, actionName flowpilot.ActionName) (*flowpilot.TransitionModel, error) {
	var transitionModel Transition

	err := flowDB.tx.Where("flow_id = ?", flowID).
		Where("action = ?", actionName).
		Order("created_at desc").
		First(&transitionModel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return transitionModel.ToFlowpilotModel(), nil
}
