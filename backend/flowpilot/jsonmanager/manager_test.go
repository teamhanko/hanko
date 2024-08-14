package jsonmanager

import (
	"github.com/tidwall/gjson"
	"reflect"
	"testing"
)

func TestDefaultJSONManager_Delete(t *testing.T) {
	type fields struct {
		data string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Valid deletion",
			fields:  fields{data: `{"key":"value"}`},
			args:    args{path: "key"},
			wantErr: false,
		},
		{
			name:    "Invalid path",
			fields:  fields{data: `{"key":"value"}`},
			args:    args{path: "nonexistent"},
			wantErr: false,
		},
		{
			name:    "Delete from empty JSON",
			fields:  fields{data: `{}`},
			args:    args{path: "key"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := &DefaultJSONManager{
				data: tt.fields.data,
			}
			if err := jm.Delete(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultJSONManager_Get(t *testing.T) {
	type fields struct {
		data string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   gjson.Result
	}{
		{
			name:   "Valid key",
			fields: fields{data: `{"key":"value"}`},
			args:   args{path: "key"},
			want: gjson.Result{
				Type:    gjson.String,
				Raw:     `"value"`,
				Str:     "value",
				Num:     0,
				Index:   7,
				Indexes: nil,
			},
		},
		{
			name:   "Nonexistent key",
			fields: fields{data: `{"key":"value"}`},
			args:   args{path: "nonexistent"},
			want:   gjson.Result{},
		},
		{
			name:   "Empty JSON",
			fields: fields{data: `{}`},
			args:   args{path: "key"},
			want:   gjson.Result{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := &DefaultJSONManager{
				data: tt.fields.data,
			}
			if got := jm.Get(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultJSONManager_Set(t *testing.T) {
	type fields struct {
		data string
	}
	type args struct {
		path  string
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Valid set",
			fields:  fields{data: `{"key":"value"}`},
			args:    args{path: "key", value: "newvalue"},
			wantErr: false,
		},
		{
			name:    "Set new key",
			fields:  fields{data: `{"key":"value"}`},
			args:    args{path: "newkey", value: "newvalue"},
			wantErr: false,
		},
		{
			name:    "Invalid path",
			fields:  fields{data: `{"key":"value"}`},
			args:    args{path: "", value: "newvalue"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := &DefaultJSONManager{
				data: tt.fields.data,
			}
			if err := jm.Set(tt.args.path, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultJSONManager_String(t *testing.T) {
	type fields struct {
		data string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Non-empty JSON",
			fields: fields{data: `{"key":"value"}`},
			want:   `{"key":"value"}`,
		},
		{
			name:   "Empty JSON",
			fields: fields{data: `{}`},
			want:   `{}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := &DefaultJSONManager{
				data: tt.fields.data,
			}
			if got := jm.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultJSONManager_Unmarshal(t *testing.T) {
	type fields struct {
		data string
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name:   "Valid JSON",
			fields: fields{data: `{"key":"value"}`},
			want:   map[string]interface{}{"key": "value"},
		},
		{
			name:   "Empty JSON",
			fields: fields{data: `{}`},
			want:   map[string]interface{}{},
		},
		{
			name:   "Invalid JSON",
			fields: fields{data: `{invalid}`},
			want:   map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := &DefaultJSONManager{
				data: tt.fields.data,
			}
			if got := jm.Unmarshal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewJSONManager(t *testing.T) {
	tests := []struct {
		name string
		want JSONManager
	}{
		{
			name: "Default manager",
			want: &DefaultJSONManager{data: "{}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewJSONManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJSONManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewJSONManagerFromString(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    JSONManager
		wantErr bool
	}{
		{
			name:    "Valid JSON",
			args:    args{data: `{"key":"value"}`},
			want:    &DefaultJSONManager{data: `{"key":"value"}`},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			args:    args{data: `{invalid}`},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewJSONManagerFromString(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJSONManagerFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJSONManagerFromString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
