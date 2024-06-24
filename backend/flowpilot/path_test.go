package flowpilot

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *defaultFlowPath
	}{
		{
			name: "construct new path",
			want: &defaultFlowPath{flowNames: []FlowName{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newFlowPath(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFlowPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	type args struct {
		root string
	}
	tests := []struct {
		name string
		args args
		want *defaultFlowPath
	}{
		{
			name: "construct path with empty string",
			args: args{root: ""},
			want: &defaultFlowPath{flowNames: []FlowName{""}},
		},
		{
			name: "construct path with root",
			args: args{root: "subflow1"},
			want: &defaultFlowPath{flowNames: []FlowName{"subflow1"}},
		},
		{
			name: "construct path with path",
			args: args{root: "subflow1.subflow2"},
			want: &defaultFlowPath{flowNames: []FlowName{"subflow1", "subflow2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newFlowPathFromPath(tt.args.root); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFlowPathFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Add(t *testing.T) {
	type fields struct {
		flowNames []FlowName
	}
	type args struct {
		flowName FlowName
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   defaultFlowPath
	}{
		{
			name:   "add to path with empty flowNames",
			fields: fields{flowNames: make([]FlowName, 0)},
			args:   args{flowName: "subflow1"},
			want:   defaultFlowPath{flowNames: []FlowName{"subflow1"}},
		},
		{
			name:   "add to path with non-empty flowNames",
			fields: fields{flowNames: []FlowName{"subflow1"}},
			args:   args{flowName: "subflow2"},
			want:   defaultFlowPath{flowNames: []FlowName{"subflow1", "subflow2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := defaultFlowPath{flowNames: tt.fields.flowNames}

			p.add(tt.args.flowName)
			if got := p; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Remove(t *testing.T) {
	type fields struct {
		flowNames []FlowName
	}
	tests := []struct {
		name   string
		fields fields
		want   defaultFlowPath
	}{
		{
			name:   "remove a flowName",
			fields: fields{flowNames: []FlowName{"subflow1", "subflow2"}},
			want:   defaultFlowPath{flowNames: []FlowName{"subflow1"}},
		},
		{
			name:   "remove until empty",
			fields: fields{flowNames: []FlowName{"subflow1"}},
			want:   defaultFlowPath{flowNames: []FlowName{}},
		},
		{
			name:   "remove when path is empty",
			fields: fields{flowNames: []FlowName{}},
			want:   defaultFlowPath{flowNames: []FlowName{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := defaultFlowPath{flowNames: tt.fields.flowNames}

			p.remove()
			if got := p; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_String(t *testing.T) {
	type fields struct {
		flowNames []FlowName
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "empty path flowNames, empty string",
			fields: fields{flowNames: make([]FlowName, 0)},
			want:   "",
		},
		{
			name:   "single path flowName",
			fields: fields{flowNames: []FlowName{"subflow1"}},
			want:   "subflow1",
		},
		{
			name:   "multiple path flowNames",
			fields: fields{flowNames: []FlowName{"subflow1", "subflow2"}},
			want:   "subflow1.subflow2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := defaultFlowPath{
				flowNames: tt.fields.flowNames,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultPath_Copy(t *testing.T) {
	t.Run("copy path and modify the original", func(t *testing.T) {
		original := &defaultFlowPath{
			flowNames: []FlowName{"a", "b"},
		}

		copied := original.copy()
		if !reflect.DeepEqual(copied, original) {
			t.Errorf("Copy() = copied version does not equal original version")
		}

		original.flowNames = append(original.flowNames, "c")
		if reflect.DeepEqual(copied, original) {
			t.Errorf("Copy() = copied version changed after original version has been modified")
		}
	})
}
