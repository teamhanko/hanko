package flowpilot

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memoryFlowDB is a minimal in-memory FlowDB for exercising flow executions.
type memoryFlowDB struct {
	flows map[uuid.UUID]FlowModel
}

func (db *memoryFlowDB) GetFlow(flowID uuid.UUID, _ uuid.UUID) (*FlowModel, error) {
	fm, ok := db.flows[flowID]
	if !ok {
		return nil, errors.New("flow not found")
	}
	return &fm, nil
}

func (db *memoryFlowDB) CreateFlow(fm FlowModel) error {
	db.flows[fm.ID] = fm
	return nil
}

func (db *memoryFlowDB) UpdateFlow(fm FlowModel) error {
	db.flows[fm.ID] = fm
	return nil
}

const (
	testStateStart StateName  = "start"
	testStateDone  StateName  = "done"
	testActionNext ActionName = "next"
)

var errTestHook = NewFlowError("test_hook_error", "Set by a hook.", http.StatusUnauthorized)

// flowErrorHook records a flow error and returns nil, the way hooks report
// errors the user can recover from.
type flowErrorHook struct{}

func (flowErrorHook) Execute(c HookExecutionContext) error {
	c.SetFlowError(errTestHook)
	return nil
}

// nextAction runs flowErrorHook and only continues when it reported no error.
type nextAction struct{}

func (nextAction) GetName() ActionName              { return testActionNext }
func (nextAction) GetDescription() string           { return "" }
func (nextAction) Initialize(InitializationContext) {}

func (nextAction) Execute(c ExecutionContext) error {
	if err := c.ExecuteHook(flowErrorHook{}); err != nil {
		return err
	}

	if flowErr := c.GetFlowError(); flowErr != nil {
		return c.Error(flowErr)
	}

	return c.Continue(testStateDone)
}

func newTestFlow(t *testing.T) Flow {
	flow, err := NewFlow("test").
		State(testStateStart, nextAction{}).
		State(testStateDone).
		State(StateName("error")).
		InitialState(testStateStart).
		ErrorState(StateName("error")).
		Build()
	require.NoError(t, err)
	return flow
}

func TestActionSeesFlowErrorSetByHook(t *testing.T) {
	db := &memoryFlowDB{flows: map[uuid.UUID]FlowModel{}}

	result, err := newTestFlow(t).Execute(db)
	require.NoError(t, err)
	require.Equal(t, testStateStart, result.GetResponse().Name)
	require.Len(t, db.flows, 1)

	var flowID uuid.UUID
	for id := range db.flows {
		flowID = id
	}

	result, err = newTestFlow(t).Execute(db,
		WithQueryParamValue(string(testActionNext)+"@"+flowID.String()),
		WithInputData(InputData{CSRFToken: result.GetResponse().CSRFToken}),
	)
	require.NoError(t, err)

	response := result.GetResponse()
	assert.Equal(t, http.StatusUnauthorized, result.GetStatus())
	assert.Equal(t, testStateStart, response.Name, "a 4xx flow error set by a hook must keep the flow in its current state")
	require.NotNil(t, response.Error)
	assert.Equal(t, "test_hook_error", response.Error.Code)
}
