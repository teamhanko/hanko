package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/dto"
)

func TestProcessTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     ClaimTemplateData
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.User.Email.Address}}",
			data: ClaimTemplateData{
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
			template: "Hello {{.User.Email.Address | printf \"%s\" | title}}",
			data: ClaimTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address: "test@example.com",
					},
				},
			},
			want:    "Hello Test@Example.Com",
			wantErr: false,
		},
		{
			name:     "template with multiple functions",
			template: "Welcome {{.User.Email.Address | printf \"%s\" | strings.ToUpper}}! Your email is {{.User.Email.Address | printf \"%s\" | strings.ToLower}}",
			data: ClaimTemplateData{
				User: &dto.UserJWT{
					Email: &dto.EmailJWT{
						Address: "Test@Example.Com",
					},
				},
			},
			want:    "Welcome TEST@EXAMPLE.COM! Your email is test@example.com",
			wantErr: false,
		},
		{
			name:     "template with conditional",
			template: "{{if .User.Email.IsVerified}}Verified{{else}}Unverified{{end}} user {{.User.Email.Address}}",
			data: ClaimTemplateData{
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
			data: ClaimTemplateData{
				User: &dto.UserJWT{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:     "empty template",
			template: "",
			data: ClaimTemplateData{
				User: &dto.UserJWT{},
			},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processTemplate(tt.template, tt.data)
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
		data    ClaimTemplateData
		want    interface{}
		wantErr bool
	}{
		{
			name:  "string template",
			value: "Hello {{.User.Email.Address}}",
			data: ClaimTemplateData{
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
			data: ClaimTemplateData{
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
			data: ClaimTemplateData{
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
			data: ClaimTemplateData{
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
			data: ClaimTemplateData{
				User: &dto.UserJWT{},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processClaimValue(tt.value, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
