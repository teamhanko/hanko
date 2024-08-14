package flowpilot

import (
	"fmt"
	"reflect"
	"testing"
)

// MockAction is a mock implementation of the Action interface.
type MockAction struct {
	Name        ActionName
	Description string
	Initialized bool
}

// NewMockAction creates a new instance of MockAction.
func NewMockAction(name ActionName, description string) *MockAction {
	return &MockAction{
		Name:        name,
		Description: description,
	}
}

// GetName returns the action name.
func (a *MockAction) GetName() ActionName {
	return a.Name
}

// GetDescription returns the action description.
func (a *MockAction) GetDescription() string {
	return a.Description
}

// Initialize initializes the action. For the mock, we just set a flag.
func (a *MockAction) Initialize(ctx InitializationContext) {
	a.Initialized = true
}

// Execute executes the action. For the mock, we just return a nil error.
func (a *MockAction) Execute(ctx ExecutionContext) error {
	if !a.Initialized {
		return fmt.Errorf("action not initialized")
	}
	return nil
}

// MockHookAction is a mock implementation of the HookAction interface.
type MockHookAction struct {
	Executed bool
	Error    error
}

// NewMockHookAction creates a new instance of MockHookAction.
func NewMockHookAction(err error) *MockHookAction {
	return &MockHookAction{
		Error: err,
	}
}

// Execute executes the hook action. It sets the Executed flag to true and returns the preset error.
func (h *MockHookAction) Execute(ctx HookExecutionContext) error {
	h.Executed = true
	return h.Error
}

func Test_defaultSubFlowBuilder_State(t *testing.T) {
	builder := NewSubFlow("testFlow")

	stateName := StateName("testState")
	action := &MockAction{}

	builder.State(stateName, action)

	sf, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() returned an error: %v", err)
	}

	// Assert the state was added
	if len(sf.(*defaultFlowBase).flow) != 1 {
		t.Errorf("State() did not add the state properly")
	}
}

func Test_defaultSubFlowBuilder_BeforeState(t *testing.T) {
	builder := NewSubFlow("testFlow")

	stateName := StateName("testState")
	hook := &MockHookAction{}

	builder.BeforeState(stateName, hook)

	sf, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() returned an error: %v", err)
	}

	// Assert the hook was added
	if len(sf.(*defaultFlowBase).beforeStateHooks[stateName]) != 1 {
		t.Errorf("BeforeState() did not add the hook properly")
	}
}

func Test_defaultSubFlowBuilder_AfterState(t *testing.T) {
	builder := NewSubFlow("testFlow")

	stateName := StateName("testState")
	hook := &MockHookAction{}

	builder.AfterState(stateName, hook)

	sf, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() returned an error: %v", err)
	}

	// Assert the hook was added
	if len(sf.(*defaultFlowBase).afterStateHooks[stateName]) != 1 {
		t.Errorf("AfterState() did not add the hook properly")
	}
}

func Test_defaultSubFlowBuilder_SubFlows(t *testing.T) {
	builder := NewSubFlow("testFlow")

	subFlowMock := &defaultFlowBase{name: "subFlow1"}

	builder.SubFlows(subFlowMock)

	sf, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() returned an error: %v", err)
	}

	// Assert the subFlow was added
	if len(sf.(*defaultFlowBase).subFlows) != 1 {
		t.Errorf("SubFlows() did not add the subFlow properly")
	}
}

func Test_defaultSubFlowBuilder_MustBuild(t *testing.T) {
	builder := NewSubFlow("testFlow")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustBuild() panicked unexpectedly: %v", r)
		}
	}()

	_ = builder.MustBuild()

	// No assertions needed; we just want to ensure no panic occurs
}

func Test_defaultSubFlowBuilder_Build(t *testing.T) {
	builder := NewSubFlow("testFlow")

	sf, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() returned an error: %v", err)
	}

	expected := &defaultFlowBase{
		name:             "testFlow",
		flow:             make(stateActions),
		subFlows:         make(SubFlows, 0),
		beforeStateHooks: make(stateHooks),
		afterStateHooks:  make(stateHooks),
	}

	if !reflect.DeepEqual(sf, expected) {
		t.Errorf("Build() returned %+v, expected %+v", sf, expected)
	}
}
