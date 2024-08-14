package flowpilot

import (
	"net/http"
	"testing"
)

func TestPredefinedFlowErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         FlowError
		wantCode    string
		wantMessage string
		wantStatus  int
	}{
		{
			name:        "ErrorTechnical",
			err:         ErrorTechnical,
			wantCode:    "technical_error",
			wantMessage: "Something went wrong.",
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:        "ErrorFlowExpired",
			err:         ErrorFlowExpired,
			wantCode:    "flow_expired_error",
			wantMessage: "The flow has expired.",
			wantStatus:  http.StatusGone,
		},
		{
			name:        "ErrorFlowDiscontinuity",
			err:         ErrorFlowDiscontinuity,
			wantCode:    "flow_discontinuity_error",
			wantMessage: "The flow can't be continued.",
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:        "ErrorOperationNotPermitted",
			err:         ErrorOperationNotPermitted,
			wantCode:    "operation_not_permitted_error",
			wantMessage: "The operation is not permitted.",
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "ErrorFormDataInvalid",
			err:         ErrorFormDataInvalid,
			wantCode:    "form_data_invalid_error",
			wantMessage: "Form data invalid.",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCode := tt.err.Code(); gotCode != tt.wantCode {
				t.Errorf("%s.Code() = %v, want %v", tt.name, gotCode, tt.wantCode)
			}
			if gotMessage := tt.err.Message(); gotMessage != tt.wantMessage {
				t.Errorf("%s.Message() = %v, want %v", tt.name, gotMessage, tt.wantMessage)
			}
			if gotStatus := tt.err.Status(); gotStatus != tt.wantStatus {
				t.Errorf("%s.Status() = %v, want %v", tt.name, gotStatus, tt.wantStatus)
			}
		})
	}
}

func TestPredefinedInputErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         InputError
		wantCode    string
		wantMessage string
	}{
		{
			name:        "ErrorValueMissing",
			err:         ErrorValueMissing,
			wantCode:    "value_missing_error",
			wantMessage: "The value is missing.",
		},
		{
			name:        "ErrorValueInvalid",
			err:         ErrorValueInvalid,
			wantCode:    "value_invalid_error",
			wantMessage: "The value is invalid.",
		},
		{
			name:        "ErrorValueTooLong",
			err:         ErrorValueTooLong,
			wantCode:    "value_too_long_error",
			wantMessage: "The value is too long.",
		},
		{
			name:        "ErrorValueTooShort",
			err:         ErrorValueTooShort,
			wantCode:    "value_too_short_error",
			wantMessage: "The value is too short.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCode := tt.err.Code(); gotCode != tt.wantCode {
				t.Errorf("%s.Code() = %v, want %v", tt.name, gotCode, tt.wantCode)
			}
			if gotMessage := tt.err.Message(); gotMessage != tt.wantMessage {
				t.Errorf("%s.Message() = %v, want %v", tt.name, gotMessage, tt.wantMessage)
			}
		})
	}
}
