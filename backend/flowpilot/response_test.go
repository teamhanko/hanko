package flowpilot

import (
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func Test_newFlowResultFromResponse(t *testing.T) {
	type args struct {
		response Response
	}
	tests := []struct {
		name string
		args args
		want FlowResult
	}{
		{
			name: "Valid response",
			args: args{
				response: Response{
					Name:    "someState",
					Status:  http.StatusOK,
					Payload: map[string]interface{}{"key": "value"},
				},
			},
			want: defaultFlowResult{
				response: Response{
					Name:    "someState",
					Status:  http.StatusOK,
					Payload: map[string]interface{}{"key": "value"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newFlowResultFromResponse(tt.args.response); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFlowResultFromResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newFlowResultFromError(t *testing.T) {
	cause := errors.New("test_cause")
	causeStr := cause.Error()

	type args struct {
		stateName StateName
		flowError FlowError
		debug     bool
	}
	tests := []struct {
		name    string
		args    args
		want    FlowResult
		wantErr bool
	}{
		{
			name: "Flow error with debug",
			args: args{
				stateName: "someState",
				flowError: &defaultFlowError{status: http.StatusBadRequest, defaultError: defaultError{code: "bad_request", message: "bad request", cause: cause}},
				debug:     true,
			},
			want: defaultFlowResult{
				response: Response{
					Name:   "someState",
					Status: http.StatusBadRequest,
					Error: &ResponseError{
						Code:     "bad_request",
						Message:  "bad request",
						Cause:    &causeStr,
						Internal: &causeStr,
					},
					Actions: ResponseActions{},
				},
			},
		},
		{
			name: "Flow error without debug",
			args: args{
				stateName: "someState",
				flowError: &defaultFlowError{status: http.StatusInternalServerError, defaultError: defaultError{code: "server_error", message: "An internal error occurred.", cause: cause}},
				debug:     false,
			},
			want: defaultFlowResult{
				response: Response{
					Name:   "someState",
					Status: http.StatusInternalServerError,
					Error: &ResponseError{
						Code:     "server_error",
						Message:  "An internal error occurred.",
						Cause:    nil,
						Internal: &causeStr,
					},
					Actions: ResponseActions{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newFlowResultFromError(tt.args.stateName, tt.args.flowError, tt.args.debug)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFlowResultFromError() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultFlowResult_GetResponse(t *testing.T) {
	type fields struct {
		response Response
	}
	tests := []struct {
		name   string
		fields fields
		want   Response
	}{
		{
			name: "Get valid response",
			fields: fields{
				response: Response{
					Name:    "someState",
					Status:  http.StatusOK,
					Payload: map[string]interface{}{"key": "value"},
				},
			},
			want: Response{
				Name:    "someState",
				Status:  http.StatusOK,
				Payload: map[string]interface{}{"key": "value"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := defaultFlowResult{
				response: tt.fields.response,
			}
			if got := r.GetResponse(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultFlowResult_GetStatus(t *testing.T) {
	type fields struct {
		response Response
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "Get status OK",
			fields: fields{
				response: Response{
					Status: http.StatusOK,
				},
			},
			want: http.StatusOK,
		},
		{
			name: "Get status Bad Request",
			fields: fields{
				response: Response{
					Status: http.StatusBadRequest,
				},
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := defaultFlowResult{
				response: tt.fields.response,
			}
			if got := r.GetStatus(); got != tt.want {
				t.Errorf("GetStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_executionResult_generateResponse(t *testing.T) {
	// Additional test setup for generateResponse would be needed
	// including creating mock/default objects for defaultFlowContext, etc.
}

func Test_executionResult_generateLinks(t *testing.T) {
	// Similar setup as above for generating links tests
}

func Test_executionResult_generateActions(t *testing.T) {
	// Similar setup as above for generating actions tests
}

func Test_executionResult_getInputSchema(t *testing.T) {
	// Similar setup as above for getting input schema tests
}

func Test_executionResult_createHref(t *testing.T) {
	// Similar setup as above for creating href tests
}
