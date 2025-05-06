package flowpilot

import (
	"github.com/gofrs/uuid"
	"net/url"
	"reflect"
	"testing"
)

func Test_createQueryParamValue(t *testing.T) {
	type args struct {
		actionName ActionName
		flowID     uuid.UUID
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Valid input",
			args: args{
				actionName: "exampleAction",
				flowID:     uuid.Nil,
			},
			want: "exampleAction@" + uuid.Nil.String(),
		},
		{
			name: "Empty action name",
			args: args{
				actionName: "",
				flowID:     uuid.Nil,
			},
			want: "@" + uuid.Nil.String(),
		},
		{
			name: "Empty flow ID",
			args: args{
				actionName: "exampleAction",
				flowID:     uuid.Nil,
			},
			want: "exampleAction@" + uuid.Nil.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createQueryParamValue(tt.args.actionName, tt.args.flowID); got != tt.want {
				t.Errorf("createQueryParamValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultQueryParam_getActionName(t *testing.T) {
	type fields struct {
		key                   string
		parsedQueryParamValue *parsedQueryParamValue
	}
	tests := []struct {
		name   string
		fields fields
		want   ActionName
	}{
		{
			name: "Valid action name",
			fields: fields{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					actionName: "testAction",
				},
			},
			want: "testAction",
		},
		{
			name: "Empty action name",
			fields: fields{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					actionName: "",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &defaultQueryParam{
				key:                   tt.fields.key,
				parsedQueryParamValue: tt.fields.parsedQueryParamValue,
			}
			if got := q.getActionName(); got != tt.want {
				t.Errorf("getActionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultQueryParam_getFlowID(t *testing.T) {
	type fields struct {
		key                   string
		parsedQueryParamValue *parsedQueryParamValue
	}
	tests := []struct {
		name   string
		fields fields
		want   uuid.UUID
	}{
		{
			name: "Valid flow ID",
			fields: fields{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					flowID: uuid.Nil,
				},
			},
			want: uuid.Nil,
		},
		{
			name: "Empty flow ID",
			fields: fields{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					flowID: uuid.Nil,
				},
			},
			want: uuid.Nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &defaultQueryParam{
				key:                   tt.fields.key,
				parsedQueryParamValue: tt.fields.parsedQueryParamValue,
			}
			if got := q.getFlowID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFlowID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultQueryParam_getKey(t *testing.T) {
	type fields struct {
		key                   string
		parsedQueryParamValue *parsedQueryParamValue
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid key",
			fields: fields{
				key: "testKey",
			},
			want: "testKey",
		},
		{
			name: "Empty key",
			fields: fields{
				key: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &defaultQueryParam{
				key:                   tt.fields.key,
				parsedQueryParamValue: tt.fields.parsedQueryParamValue,
			}
			if got := q.getKey(); got != tt.want {
				t.Errorf("getKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultQueryParam_getURLValues(t *testing.T) {
	type fields struct {
		key                   string
		parsedQueryParamValue *parsedQueryParamValue
	}
	tests := []struct {
		name   string
		fields fields
		want   url.Values
	}{
		{
			name: "Valid URL values",
			fields: fields{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					actionName: "testAction",
					flowID:     uuid.Nil,
				},
			},
			want: url.Values{
				"testKey": []string{"testAction@" + uuid.Nil.String()},
			},
		},
		{
			name: "Empty key",
			fields: fields{
				key: "",
				parsedQueryParamValue: &parsedQueryParamValue{
					actionName: "testAction",
					flowID:     uuid.Nil,
				},
			},
			want: url.Values{
				"": []string{"testAction@" + uuid.Nil.String()},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &defaultQueryParam{
				key:                   tt.fields.key,
				parsedQueryParamValue: tt.fields.parsedQueryParamValue,
			}
			if got := q.getURLValues(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getURLValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultQueryParam_getValue(t *testing.T) {
	type fields struct {
		key                   string
		parsedQueryParamValue *parsedQueryParamValue
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Valid value",
			fields: fields{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					actionName: "testAction",
					flowID:     uuid.Nil,
				},
			},
			want: "testAction@" + uuid.Nil.String(),
		},
		{
			name: "Empty value",
			fields: fields{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					actionName: "testAction",
					flowID:     uuid.Nil,
				},
			},
			want: "testAction@" + uuid.Nil.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &defaultQueryParam{
				key:                   tt.fields.key,
				parsedQueryParamValue: tt.fields.parsedQueryParamValue,
			}
			if got := q.getValue(); got != tt.want {
				t.Errorf("getValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newQueryParam(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    queryParam
		wantErr bool
	}{
		{
			name: "Valid query param",
			args: args{
				key:   "testKey",
				value: "testAction@" + uuid.Nil.String(),
			},
			want: &defaultQueryParam{
				key: "testKey",
				parsedQueryParamValue: &parsedQueryParamValue{
					actionName: "testAction",
					flowID:     uuid.Nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid format",
			args: args{
				key:   "testKey",
				value: "invalidFormat",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty value",
			args: args{
				key:   "testKey",
				value: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newQueryParam(tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("newQueryParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newQueryParam() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseQueryParamValue(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    *parsedQueryParamValue
		wantErr bool
	}{
		{
			name: "Valid input",
			args: args{
				value: "testAction@" + uuid.Nil.String(),
			},
			want: &parsedQueryParamValue{
				actionName: "testAction",
				flowID:     uuid.Nil,
			},
			wantErr: false,
		},
		{
			name: "Invalid format",
			args: args{
				value: "invalidFormat",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty value",
			args: args{
				value: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseQueryParamValue(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseQueryParamValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseQueryParamValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}
