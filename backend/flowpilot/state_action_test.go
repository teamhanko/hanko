package flowpilot

import (
	"reflect"
	"testing"
)

func Test_defaultActionDetail_getAction(t *testing.T) {
	type fields struct {
		action   Action
		flowName FlowName
	}
	tests := []struct {
		name   string
		fields fields
		want   Action
	}{
		{
			name: "Valid action",
			fields: fields{
				action: &mockAction{name: "action1"},
			},
			want: &mockAction{name: "action1"},
		},
		{
			name: "Nil action",
			fields: fields{
				action: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := &defaultActionDetail{
				action:   tt.fields.action,
				flowName: tt.fields.flowName,
			}
			if got := ad.getAction(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultActionDetail_getFlowName(t *testing.T) {
	type fields struct {
		action   Action
		flowName FlowName
	}
	tests := []struct {
		name   string
		fields fields
		want   FlowName
	}{
		{
			name: "Valid flow name",
			fields: fields{
				flowName: "flow1",
			},
			want: "flow1",
		},
		{
			name: "Empty flow name",
			fields: fields{
				flowName: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := &defaultActionDetail{
				action:   tt.fields.action,
				flowName: tt.fields.flowName,
			}
			if got := ad.getFlowName(); got != tt.want {
				t.Errorf("getFlowName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Mock implementations for testing purposes
type mockAction struct {
	name string
}

func (ma *mockAction) GetName() ActionName {
	return ActionName(ma.name)
}

func (ma *mockAction) GetDescription() string {
	return "mock description"
}

func (ma *mockAction) Initialize(ctx InitializationContext) {}

func (ma *mockAction) Execute(ctx ExecutionContext) error {
	return nil
}
