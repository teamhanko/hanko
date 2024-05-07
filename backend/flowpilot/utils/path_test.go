package utils

import (
	"reflect"
	"testing"
)

func TestNewPath(t *testing.T) {
	type args struct {
		root string
	}
	tests := []struct {
		name string
		args args
		want Path
	}{
		{
			name: "construct path with empty string",
			args: args{root: ""},
			want: Path{fragments: []string{""}},
		},
		{
			name: "construct path with root",
			args: args{root: "subflow1"},
			want: Path{fragments: []string{"subflow1"}},
		},
		{
			name: "construct path with path",
			args: args{root: "subflow1.subflow2"},
			want: Path{fragments: []string{"subflow1", "subflow2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPath(tt.args.root); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Add(t *testing.T) {
	type fields struct {
		fragments []string
	}
	type args struct {
		fragment string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Path
	}{
		{
			name:   "add to path with empty fragments",
			fields: fields{fragments: make([]string, 0)},
			args:   args{fragment: "subflow1"},
			want:   Path{fragments: []string{"subflow1"}},
		},
		{
			name:   "add to path with non-empty fragments",
			fields: fields{fragments: []string{"subflow1"}},
			args:   args{fragment: "subflow2"},
			want:   Path{fragments: []string{"subflow1", "subflow2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Path{fragments: tt.fields.fragments}

			p.Add(tt.args.fragment)
			if got := p; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Remove(t *testing.T) {
	type fields struct {
		fragments []string
	}
	tests := []struct {
		name   string
		fields fields
		want   Path
	}{
		{
			name:   "remove a fragment",
			fields: fields{fragments: []string{"subflow1", "subflow2"}},
			want:   Path{fragments: []string{"subflow1"}},
		},
		{
			name:   "remove until empty",
			fields: fields{fragments: []string{"subflow1"}},
			want:   Path{fragments: []string{}},
		},
		{
			name:   "remove when path is empty",
			fields: fields{fragments: []string{}},
			want:   Path{fragments: []string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Path{fragments: tt.fields.fragments}

			p.Remove()
			if got := p; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_String(t *testing.T) {
	type fields struct {
		fragments []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "empty path fragments, empty string",
			fields: fields{fragments: make([]string, 0)},
			want:   "",
		},
		{
			name:   "single path fragment",
			fields: fields{fragments: []string{"subflow1"}},
			want:   "subflow1",
		},
		{
			name:   "multiple path fragments",
			fields: fields{fragments: []string{"subflow1", "subflow2"}},
			want:   "subflow1.subflow2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Path{
				fragments: tt.fields.fragments,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
