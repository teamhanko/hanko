package session

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/dto"
)

func TestProcessJWTTemplate(t *testing.T) {
	tests := []struct {
		name           string
		claims         map[string]interface{}
		user           dto.UserJWT
		expectedClaims json.RawMessage
		wantErr        bool
	}{
		{
			name: "should process static claims with basic and complex types",
			claims: map[string]interface{}{
				"static_str":      "foo",
				"static_num":      123,
				"static_bool":     true,
				"static_str_arr":  []string{"a", "b", "c"},
				"static_num_arr":  []int{1, 2, 3},
				"static_bool_arr": []bool{true, false, true},
				"static_mix_arr":  []interface{}{"a", 1, true},
				"static_obj": struct {
					A string
					B []string
					C struct {
						D []interface{}
					}
				}{
					A: "a",
					B: []string{"a", "b", "c"},
					C: struct {
						D []interface{}
					}{
						D: []interface{}{"a", 1, true},
					},
				},
			},
			user: dto.UserJWT{
				Email: &dto.EmailJWT{
					Address: "test@example.com",
				},
			},
			expectedClaims: json.RawMessage(`{
				"static_str": "foo",
				"static_num": 123,
				"static_bool": true,
				"static_str_arr":  ["a", "b", "c"],
				"static_num_arr":  [1, 2, 3],
				"static_bool_arr": [true, false, true],
				"static_mix_arr":  ["a", 1, true],
				"static_obj": {
					"A": "a",
					"B": ["a", "b", "c"],
					"C": {
						"D": ["a", 1, true]
					}
				}
			}`),
		},
		{
			name: "should process access to top level user context object",
			claims: map[string]interface{}{
				"user": "{{ .User }}",
			},
			user: dto.UserJWT{
				Email: &dto.EmailJWT{
					Address:    "test@example.com",
					IsVerified: true,
					IsPrimary:  true,
				},
				Username: "test_user",
				UserID:   "48986f51-d9c8-4f22-89e9-fb7fab959399",
			},
			expectedClaims: json.RawMessage(`{
				"user": {
					"email": {
						"address": "test@example.com",
						"is_verified": true,
						"is_primary": true
					},
					"user_id": "48986f51-d9c8-4f22-89e9-fb7fab959399",
					"username": "test_user"
				}
			}`),
		},
		{
			name: "should process access to top level fields of user context object",
			claims: map[string]interface{}{
				"user_id":  "{{ .User.UserID }}",
				"username": "{{ .User.Username }}",
				"email":    "{{ .User.Email }}",
				"metadata": "{{ .User.Metadata }}",
			},
			user: dto.UserJWT{
				UserID:   "48986f51-d9c8-4f22-89e9-fb7fab959399",
				Username: "test_user",
				Email: &dto.EmailJWT{
					Address:    "test@example.com",
					IsVerified: true,
					IsPrimary:  true,
				},
				Metadata: dto.NewMetadataJWT(
					json.RawMessage(`{"public_key": "public_value"}`),
					json.RawMessage(`{"unsafe_key": "unsafe_value"}`),
				),
			},
			expectedClaims: json.RawMessage(`{
				"user_id": "48986f51-d9c8-4f22-89e9-fb7fab959399",
				"username": "test_user",
				"email": {
					"address": "test@example.com",
					"is_verified": true,
					"is_primary": true
				},
				"metadata": {
					"public_metadata": {
						"public_key": "public_value"
					},
					"unsafe_metadata": {
						"unsafe_key": "unsafe_value"
					}
				}
			}`),
		},
		{
			name: "should process more complex go templates",
			claims: map[string]interface{}{
				"greeting":                     "Hello {{ .User.Email.Address }}",
				"greeting_pipelined":           "Hello {{ .User.Email.Address | printf \"%s\" }}",
				"verification_msg_conditional": "{{if .User.Email.IsVerified}}Verified{{else}}Unverified{{end}} user {{.User.Email.Address}}",
				"verification_msg_conditional_pretty": `
					{{- if .User.Email.IsVerified -}}
						Verified
					{{- else -}}
						Unverified
					{{ end }} user {{ .User.Email.Address }}`,
			},
			user: dto.UserJWT{
				Email: &dto.EmailJWT{
					Address:    "test@example.com",
					IsVerified: true,
					IsPrimary:  true,
				},
				Username: "test_user",
				UserID:   "48986f51-d9c8-4f22-89e9-fb7fab959399",
			},
			expectedClaims: json.RawMessage(`{
				"greeting": "Hello test@example.com",
				"greeting_pipelined": "Hello test@example.com",
				"verification_msg_conditional": "Verified user test@example.com",
				"verification_msg_conditional_pretty": "Verified user test@example.com"
			}`),
		},
		{
			name: "should ignore entries with invalid templates or processing errors",
			claims: map[string]interface{}{
				"valid_template":                        "Hello {{ .User.Username }}",
				"invalid_template":                      "Hello {{ .User.Username }",
				"non_existing_field_on_context_data":    "Hello {{ .User.Surname }}",
				"non_existing_function_on_context_data": `Hello {{ .User.Metadata.Private "private_key" }}`,
			},
			user: dto.UserJWT{
				Email: &dto.EmailJWT{
					Address:    "test@example.com",
					IsVerified: true,
					IsPrimary:  true,
				},
				Username: "test_user",
				UserID:   "48986f51-d9c8-4f22-89e9-fb7fab959399",
				Metadata: dto.NewMetadataJWT(
					json.RawMessage(`{"public_key": "public_value"}`),
					json.RawMessage(`{"unsafe_key": "unsafe_value"}`),
				),
			},
			expectedClaims: json.RawMessage(`{
				"valid_template": "Hello test_user"
			}`),
			wantErr: true,
		},
		{
			name: "should process access to metadata",
			claims: map[string]interface{}{
				"display_name":                          `{{ .User.Metadata.Public "display_name" }}`,
				"favorite_games":                        `{{ .User.Metadata.Public "favorite_games" }}`,
				"favorite_games_with_playtime_over_100": `{{ .User.Metadata.Public "favorite_games.#(playtime_hours>100)" }}`,
				"favorite_genres":                       `{{ .User.Metadata.Public "favorite_games.#.genre"}}`,
				"ui_theme":                              `{{ .User.Metadata.Unsafe "ui_theme" }}`,
			},
			user: dto.UserJWT{
				Metadata: dto.NewMetadataJWT(
					json.RawMessage(`{
						"display_name": "GamerDude",
						"favorite_games": [
							{
								"name": "Legends of Valor",
								"genre": "RPG",
								"playtime_hours": 142.3
							},
							{
								"name": "Space Raiders",
								"genre": "Sci-Fi Shooter",
								"playtime_hours": 87.6
							}
						]
					}`),
					json.RawMessage(`{
						"ui_theme": "dark"
					}`),
				),
			},
			expectedClaims: json.RawMessage(`{
				"display_name": "GamerDude",
				"favorite_games": [
					{
						"name": "Legends of Valor",
						"genre": "RPG",
						"playtime_hours": 142.3
					},
					{
						"name": "Space Raiders",
						"genre": "Sci-Fi Shooter",
						"playtime_hours": 87.6
					}
				],
				"favorite_games_with_playtime_over_100": {
					"name": "Legends of Valor",
					"genre": "RPG",
					"playtime_hours": 142.3
				},
				"favorite_genres": ["RPG", "Sci-Fi Shooter"],
				"ui_theme": "dark"
			}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.New()
			err := ProcessJWTTemplate(token, tt.claims, tt.user)
			assert.NoError(t, err)

			privateClaims := token.PrivateClaims()
			privateClaimsBytes, err := json.Marshal(privateClaims)
			assert.NoError(t, err)

			require.True(t, gjson.ValidBytes(tt.expectedClaims))
			require.True(t, gjson.ValidBytes(privateClaimsBytes))

			assert.Equal(
				t,
				gjson.GetBytes(tt.expectedClaims, `@pretty:{"sortKeys":true}`).String(),
				gjson.GetBytes(privateClaimsBytes, `@pretty:{"sortKeys":true}`).String(),
			)
		})
	}
}
