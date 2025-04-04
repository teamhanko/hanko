package session

import (
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/dto"
)

func TestProcessTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     JWTTemplateData
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.User.Email.Address}}",
			data: JWTTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address: "test@example.com",
					},
				},
			},
			want:    "Hello test@example.com",
			wantErr: false,
		},
		{
			name:     "template with pipelining",
			template: "Hello {{.User.Email.Address | printf \"%s\" }}",
			data: JWTTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address: "test@example.com",
					},
				},
			},
			want:    "Hello test@example.com",
			wantErr: false,
		},
		{
			name:     "template with conditional",
			template: "{{if .User.Email.IsVerified}}Verified{{else}}Unverified{{end}} user {{.User.Email.Address}}",
			data: JWTTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address:    "test@example.com",
						IsVerified: true,
					},
				},
			},
			want:    "Verified user test@example.com",
			wantErr: false,
		},
		{
			name:     "invalid template",
			template: "Hello {{.InvalidField}}",
			data: JWTTemplateData{
				User: &dto.UserJWT{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:     "empty template",
			template: "",
			data: JWTTemplateData{
				User: &dto.UserJWT{},
			},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseClaimTemplateValue(tt.template, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProcessClaimValue(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		data    JWTTemplateData
		want    interface{}
		wantErr bool
	}{
		{
			name:  "string template",
			value: "Hello {{.User.Email.Address}}",
			data: JWTTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address: "test@example.com",
					},
				},
			},
			want:    "Hello test@example.com",
			wantErr: false,
		},
		{
			name: "nested map with templates",
			value: map[string]interface{}{
				"greeting": "Hello {{.User.Email.Address}}",
				"nested": map[string]interface{}{
					"message": "Welcome {{.User.Email.Address}}",
				},
			},
			data: JWTTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address: "test@example.com",
					},
				},
			},
			want: map[string]interface{}{
				"greeting": "Hello test@example.com",
				"nested": map[string]interface{}{
					"message": "Welcome test@example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "slice with templates",
			value: []interface{}{
				"Hello {{.User.Email.Address}}",
				map[string]interface{}{
					"message": "Welcome {{.User.Email.Address}}",
				},
			},
			data: JWTTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address: "test@example.com",
					},
				},
			},
			want: []interface{}{
				"Hello test@example.com",
				map[string]interface{}{
					"message": "Welcome test@example.com",
				},
			},
			wantErr: false,
		},
		{
			name:  "non-string primitive",
			value: 42,
			data: JWTTemplateData{
				User: &dto.UserJWT{},
			},
			want:    42,
			wantErr: false,
		},
		{
			name: "invalid template in map",
			value: map[string]interface{}{
				"message": "Hello {{.InvalidField}}",
			},
			data: JWTTemplateData{
				User: &dto.UserJWT{},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processClaimTemplate(tt.value, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProcessClaimTemplate(t *testing.T) {
	tests := []struct {
		name           string
		claims         map[string]interface{}
		user           dto.UserJWT
		expectedClaims map[string]interface{}
	}{
		{
			name: "successful claim processing",
			claims: map[string]interface{}{
				"email":         "{{.User.Email.Address}}",
				"verified":      "{{.User.Email.IsVerified}}",
				"static_string": "static-value",
				"static_bool":   false,
			},
			user: dto.UserJWT{
				Email: &dto.EmailJWT{
					Address:    "test@example.com",
					IsVerified: true,
				},
			},
			expectedClaims: map[string]interface{}{
				"email":         "test@example.com",
				"verified":      true,
				"static_string": "static-value",
				"static_bool":   false,
			},
		},
		{
			name: "partial claim processing with errors",
			claims: map[string]interface{}{
				"valid":   "{{.User.Email.Address}}",
				"invalid": "{{.InvalidField}}",
				"static":  "static-value",
			},
			user: dto.UserJWT{
				Email: &dto.EmailJWT{
					Address: "test@example.com",
				},
			},
			expectedClaims: map[string]interface{}{
				"valid":  "test@example.com",
				"static": "static-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.New()
			err := ProcessJWTTemplate(token, tt.claims, tt.user)
			assert.NoError(t, err)

			// Verify each expected claim
			for key, expectedValue := range tt.expectedClaims {
				value, exists := token.Get(key)
				assert.True(t, exists, "claim %s should exist", key)
				assert.Equal(t, expectedValue, value, "claim %s should have correct value", key)
			}

			// For the error case, verify the invalid claim was not set
			if tt.name == "partial claim processing with errors" {
				_, exists := token.Get("invalid")
				assert.False(t, exists, "invalid claim should not be set")
			}
		})
	}
}
