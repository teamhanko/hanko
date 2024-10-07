package flowpilot

import (
	"errors"
	"reflect"
	"testing"
)

func Test_defaultError_Code(t *testing.T) {
	tests := []struct {
		name string
		err  defaultError
		want string
	}{
		{
			name: "Returns correct code",
			err:  defaultError{code: "error_code"},
			want: "error_code",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Code(); got != tt.want {
				t.Errorf("defaultError.Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultError_Message(t *testing.T) {
	tests := []struct {
		name string
		err  defaultError
		want string
	}{
		{
			name: "Returns correct message",
			err:  defaultError{message: "error message"},
			want: "error message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Message(); got != tt.want {
				t.Errorf("defaultError.Message() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultError_Unwrap(t *testing.T) {
	cause := errors.New("original cause")
	tests := []struct {
		name string
		err  defaultError
		want error
	}{
		{
			name: "Unwraps correctly",
			err:  defaultError{cause: cause},
			want: cause,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Unwrap(); got != tt.want {
				t.Errorf("defaultError.Unwrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  defaultError
		want string
	}{
		{
			name: "Returns correct error string",
			err:  defaultError{errorText: "error_code - error message"},
			want: "error_code - error message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("defaultError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultError_toResponseError(t *testing.T) {
	cause := "cause of the error"
	tests := []struct {
		name  string
		err   defaultError
		debug bool
		want  *ResponseError
	}{
		{
			name: "Converts to ResponseError with cause in debug mode",
			err: defaultError{
				cause:   errors.New(cause),
				code:    "error_code",
				message: "error message",
			},
			debug: true,
			want: &ResponseError{
				Code:     "error_code",
				Message:  "error message",
				Cause:    &cause,
				Internal: &cause,
			},
		},
		{
			name: "Converts to ResponseError without cause in non-debug mode",
			err: defaultError{
				cause:   errors.New(cause),
				code:    "error_code",
				message: "error message",
			},
			debug: false,
			want: &ResponseError{
				Code:     "error_code",
				Message:  "error message",
				Internal: &cause,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.toResponseError(tt.debug); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultError.toResponseError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createErrorText(t *testing.T) {
	type args struct {
		code    string
		message string
		cause   error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Creates error text with cause",
			args: args{
				code:    "error_code",
				message: "error message",
				cause:   errors.New("original cause"),
			},
			want: "error_code - error message: original cause",
		},
		{
			name: "Creates error text without cause",
			args: args{
				code:    "error_code",
				message: "error message",
				cause:   nil,
			},
			want: "error_code - error message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createErrorText(tt.args.code, tt.args.message, tt.args.cause); got != tt.want {
				t.Errorf("createErrorText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newFlowErrorWithCause(t *testing.T) {
	cause := errors.New("cause of the error")
	tests := []struct {
		name    string
		code    string
		message string
		status  int
		cause   error
		want    FlowError
	}{
		{
			name:    "Creates FlowError with cause",
			code:    "error_code",
			message: "error message",
			status:  400,
			cause:   cause,
			want: &defaultFlowError{
				defaultError: defaultError{
					cause:     cause,
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message: cause of the error",
				},
				status: 400,
			},
		},
		{
			name:    "Creates FlowError without cause",
			code:    "error_code",
			message: "error message",
			status:  400,
			cause:   nil,
			want: &defaultFlowError{
				defaultError: defaultError{
					cause:     nil,
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message",
				},
				status: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newFlowErrorWithCause(tt.code, tt.message, tt.status, tt.cause); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFlowErrorWithCause() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultFlowError_Status(t *testing.T) {
	tests := []struct {
		name string
		err  defaultFlowError
		want int
	}{
		{
			name: "Returns correct status",
			err:  defaultFlowError{status: 400},
			want: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Status(); got != tt.want {
				t.Errorf("defaultFlowError.Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultFlowError_Wrap(t *testing.T) {
	cause := errors.New("cause of the error")
	tests := []struct {
		name string
		err  defaultFlowError
		wrap error
		want FlowError
	}{
		{
			name: "Wraps error with new cause",
			err: defaultFlowError{
				defaultError: defaultError{
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message",
				},
				status: 400,
			},
			wrap: cause,
			want: &defaultFlowError{
				defaultError: defaultError{
					cause:     cause,
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message: cause of the error",
				},
				status: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Wrap(tt.wrap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultFlowError.Wrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newInputErrorWithCause(t *testing.T) {
	cause := errors.New("cause of the error")
	tests := []struct {
		name    string
		code    string
		message string
		cause   error
		want    InputError
	}{
		{
			name:    "Creates InputError with cause",
			code:    "error_code",
			message: "error message",
			cause:   cause,
			want: &defaultInputError{
				defaultError: defaultError{
					cause:     cause,
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message: cause of the error",
				},
			},
		},
		{
			name:    "Creates InputError without cause",
			code:    "error_code",
			message: "error message",
			cause:   nil,
			want: &defaultInputError{
				defaultError: defaultError{
					cause:     nil,
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newInputErrorWithCause(tt.code, tt.message, tt.cause); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newInputErrorWithCause() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultInputError_Wrap(t *testing.T) {
	cause := errors.New("cause of the error")
	tests := []struct {
		name string
		err  defaultInputError
		wrap error
		want InputError
	}{
		{
			name: "Wraps error with new cause",
			err: defaultInputError{
				defaultError: defaultError{
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message",
				},
			},
			wrap: cause,
			want: &defaultInputError{
				defaultError: defaultError{
					cause:     cause,
					code:      "error_code",
					message:   "error message",
					errorText: "error_code - error message: cause of the error",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Wrap(tt.wrap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultInputError.Wrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createMustBeOneOfError(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   InputError
	}{
		{
			name:   "Creates InputError for must be one of",
			values: []string{"value1", "value2", "value3"},
			want: &defaultInputError{
				defaultError: defaultError{
					code:      "value_invalid_error",
					message:   "The value is invalid. Must be one of: value1,value2,value3",
					errorText: "value_invalid_error - The value is invalid. Must be one of: value1,value2,value3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createMustBeOneOfError(tt.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createMustBeOneOfError() = %v, want %v", got, tt.want)
			}
		})
	}
}
