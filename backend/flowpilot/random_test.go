package flowpilot

import (
	"testing"
)

func Test_generateRandomString(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "Generate 10 character string",
			length:  10,
			wantErr: false,
		},
		{
			name:    "Generate 0 character string",
			length:  0,
			wantErr: false,
		},
		{
			name:    "Generate 100 character string",
			length:  100,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateRandomString(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateRandomString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.length {
				t.Errorf("generateRandomString() got length = %v, want length %v", len(got), tt.length)
			}
			for _, char := range got {
				if !contains(letters, char) {
					t.Errorf("generateRandomString() contains invalid character = %v", char)
				}
			}
		})
	}
}

func contains(str string, char rune) bool {
	for _, c := range str {
		if c == char {
			return true
		}
	}
	return false
}

func Test_assertAvailablePRNG(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("assertAvailablePRNG() panicked: %v", r)
		}
	}()

	assertAvailablePRNG() // This should not panic under normal conditions
}
