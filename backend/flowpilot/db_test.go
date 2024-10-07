package flowpilot

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// Mock for FlowDB
type MockFlowDB struct {
	mock.Mock
}

func (m *MockFlowDB) GetFlow(flowID uuid.UUID) (*FlowModel, error) {
	args := m.Called(flowID)
	return args.Get(0).(*FlowModel), args.Error(1)
}

func (m *MockFlowDB) CreateFlow(flowModel FlowModel) error {
	args := m.Called(flowModel)
	return args.Error(0)
}

func (m *MockFlowDB) UpdateFlow(flowModel FlowModel) error {
	args := m.Called(flowModel)
	return args.Error(0)
}

// Test for createFlowWithParam
func Test_defaultFlowDBWrapper_createFlowWithParam(t *testing.T) {
	mockDB := new(MockFlowDB)
	wrapper := wrapDB(mockDB)

	now := time.Now().UTC()

	mockDB.On("CreateFlow", mock.MatchedBy(func(fm FlowModel) bool {
		return fm.Data == "test data" &&
			fm.CSRFToken == "csrf-token" &&
			fm.Version == 0 &&
			fm.ExpiresAt.Equal(time.Date(2024, time.August, 15, 18, 55, 8, 0, time.UTC)) &&
			// Allow a margin for timestamp comparison
			fm.CreatedAt.After(now.Add(-time.Second)) && fm.CreatedAt.Before(now.Add(time.Second)) &&
			fm.UpdatedAt.After(now.Add(-time.Second)) && fm.UpdatedAt.Before(now.Add(time.Second))
	})).Return(nil)

	params := flowCreationParam{
		data:      "test data",
		csrfToken: "csrf-token",
		expiresAt: time.Date(2024, time.August, 15, 18, 55, 8, 0, time.UTC),
	}

	_, err := wrapper.createFlowWithParam(params)
	if err != nil {
		t.Errorf("createFlowWithParam() error = %v", err)
	}

	mockDB.AssertExpectations(t)
}

// Test for updateFlowWithParam
func Test_defaultFlowDBWrapper_updateFlowWithParam(t *testing.T) {
	mockDB := new(MockFlowDB)
	wrapper := wrapDB(mockDB)

	now := time.Now().UTC()
	fakeUUID, _ := uuid.NewV4()

	mockDB.On("UpdateFlow", mock.MatchedBy(func(fm FlowModel) bool {
		return fm.ID == fakeUUID && // Match with the UUID used
			fm.Data == "updated data" &&
			fm.CSRFToken == "updated-token" &&
			fm.Version == 1 &&
			fm.ExpiresAt.Equal(time.Date(2024, time.August, 16, 18, 55, 8, 0, time.UTC)) &&
			fm.CreatedAt.Equal(time.Date(2024, time.August, 14, 16, 55, 8, 0, time.UTC)) &&
			fm.UpdatedAt.After(now.Add(-time.Second)) && fm.UpdatedAt.Before(now.Add(time.Second))
	})).Return(nil)

	params := flowUpdateParam{
		flowID:    fakeUUID,
		data:      "updated data",
		version:   1,
		csrfToken: "updated-token",
		expiresAt: time.Date(2024, time.August, 16, 18, 55, 8, 0, time.UTC),
		createdAt: time.Date(2024, time.August, 14, 16, 55, 8, 0, time.UTC),
	}

	_, err := wrapper.updateFlowWithParam(params)
	if err != nil {
		t.Errorf("updateFlowWithParam() error = %v", err)
	}

	mockDB.AssertExpectations(t)
}
