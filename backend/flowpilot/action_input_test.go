package flowpilot

import (
	"reflect"
	"testing"

	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

func Test_newActionInput(t *testing.T) {
	tests := []struct {
		name string
		want actionInput
	}{
		{
			name: "Default instance",
			want: jsonmanager.NewJSONManager(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newActionInput(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newActionInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newActionInputFromInputData(t *testing.T) {
	type args struct {
		data InputData
	}
	tests := []struct {
		name    string
		args    args
		want    actionInput
		wantErr bool
	}{
		{
			name: "Valid InputData",
			args: args{
				data: InputData{
					InputDataMap: map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			want: func() actionInput {
				data := `{"key1":"value1","key2":"value2"}`
				m, _ := jsonmanager.NewJSONManagerFromString(data)
				return m
			}(),
			wantErr: false,
		},
		{
			name: "Empty InputData",
			args: args{
				data: InputData{
					InputDataMap: map[string]interface{}{},
				},
			},
			want: func() actionInput {
				data := `{}`
				m, _ := jsonmanager.NewJSONManagerFromString(data)
				return m
			}(),
			wantErr: false,
		},
		{
			name: "Invalid JSON InputData",
			args: args{
				data: InputData{
					InputDataMap: map[string]interface{}{
						"key1": func() {}, // Functions cannot be marshalled into JSON
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newActionInputFromInputData(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("newActionInputFromInputData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newActionInputFromInputData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
