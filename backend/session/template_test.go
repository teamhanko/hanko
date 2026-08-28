package session

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/dto"
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
				Username:   "test_user",
				UserID:     "48986f51-d9c8-4f22-89e9-fb7fab959399",
				Name:       "John Doe",
				GivenName:  "John",
				FamilyName: "Doe",
				Picture:    "https://example.com/avatar.png",
			},
			expectedClaims: json.RawMessage(`{
				"user": {
					"email": {
						"address": "test@example.com",
						"is_verified": true,
						"is_primary": true
					},
					"family_name": "Doe",
					"given_name": "John",
					"name": "John Doe",
					"picture": "https://example.com/avatar.png",
					"user_id": "48986f51-d9c8-4f22-89e9-fb7fab959399",
					"username": "test_user"
				}
			}`),
		},
		{
			name: "should process access to top level fields of user context object",
			claims: map[string]interface{}{
				"user_id":    "{{ .User.UserID }}",
				"username":   "{{ .User.Username }}",
				"email":      "{{ .User.Email }}",
				"metadata":   "{{ .User.Metadata }}",
				"full_name":  "{{ .User.Name }}",
				"first_name": "{{ .User.GivenName }}",
				"last_name":  "{{ .User.FamilyName }}",
				"avatar":     "{{ .User.Picture }}",
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
				Name:       "John Doe",
				GivenName:  "John",
				FamilyName: "Doe",
				Picture:    "https://example.com/avatar.png",
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
				},
				"full_name": "John Doe",
				"first_name": "John",
				"last_name": "Doe",
				"avatar": "https://example.com/avatar.png"
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
		{
			name: "should process access to custom claims",
			claims: map[string]interface{}{
				"custom_claims_whole": "{{ .User.CustomClaims }}",
				"matriculation_number": `{{ .User.CustomClaims "matriculation_number" }}`,
				"age":                  `{{ .User.CustomClaims "age" }}`,
				"is_staff":             `{{ .User.CustomClaims "is_staff" }}`,
				"affiliation":          `{{ .User.CustomClaims "affiliation" }}`,
			},
			user: dto.UserJWT{}.WithCustomClaims(json.RawMessage(`{
				"matriculation_number": "12345",
				"age": 29,
				"is_staff": true,
				"affiliation": ["student", "staff"]
			}`)),
			expectedClaims: json.RawMessage(`{
				"custom_claims_whole": {
					"matriculation_number": "12345",
					"age": 29,
					"is_staff": true,
					"affiliation": ["student", "staff"]
				},
				"matriculation_number": "12345",
				"age": 29,
				"is_staff": true,
				"affiliation": ["student", "staff"]
			}`),
		},
		{
			name: "should preserve a custom claim's real type only for a bare accessor, not glued to other text",
			claims: map[string]interface{}{
				"age_bare":      `{{ .User.CustomClaims "age" }}`,
				"age_composite": `Age: {{ .User.CustomClaims "age" }}`,
			},
			user: dto.UserJWT{}.WithCustomClaims(json.RawMessage(`{"age": 29}`)),
			expectedClaims: json.RawMessage(`{
				"age_bare": 29,
				"age_composite": "Age: 29"
			}`),
		},
		{
			name: "should return an empty string for a claim key that isn't in the user's data",
			claims: map[string]interface{}{
				"missing_claim": `{{ .User.CustomClaims "not_a_real_claim" }}`,
			},
			user:           dto.UserJWT{}.WithCustomClaims(json.RawMessage(`{"age": 29}`)),
			expectedClaims: json.RawMessage(`{"missing_claim": ""}`),
		},
		{
			name: "should return an empty string for a user with no custom claims at all",
			claims: map[string]interface{}{
				"bare":  "{{ .User.CustomClaims }}",
				"named": `{{ .User.CustomClaims "age" }}`,
			},
			user: dto.UserJWT{
				Username: "test_user",
			},
			expectedClaims: json.RawMessage(`{
				"bare": "",
				"named": ""
			}`),
		},
		{
			// CustomClaims used to always return a Go string, so a false boolean claim rendered
			// as the non-empty text "false" - which {{if}} treats as truthy, taking the wrong
			// branch. CustomClaims now returns the value's real Go type (see its doc comment in
			// dto/user.go), so {{if}}/{{and}}/{{or}}/{{not}}/eq all see the actual bool/number/
			// slice and behave correctly - not just a bare claim template's own output type.
			name: "should evaluate a custom claim's real type correctly inside conditionals, not just as a bare value",
			claims: map[string]interface{}{
				"is_staff_if":       `{{if .User.CustomClaims "is_staff"}}yes{{else}}no{{end}}`,
				"is_staff_not":      `{{if not (.User.CustomClaims "is_staff")}}yes{{else}}no{{end}}`,
				"affiliation_empty": `{{if .User.CustomClaims "affiliation"}}has_some{{else}}none{{end}}`,
				"age_eq":            `{{if eq (.User.CustomClaims "age") 29.0}}matched{{else}}no_match{{end}}`,
				"missing_claim_if":  `{{if .User.CustomClaims "not_a_real_claim"}}yes{{else}}no{{end}}`,
			},
			user: dto.UserJWT{}.WithCustomClaims(json.RawMessage(`{
				"is_staff": false,
				"affiliation": [],
				"age": 29
			}`)),
			expectedClaims: json.RawMessage(`{
				"is_staff_if": "no",
				"is_staff_not": "yes",
				"affiliation_empty": "none",
				"age_eq": "matched",
				"missing_claim_if": "no"
			}`),
		},
		{
			name: "should fall back to a stringified array/object for a custom claim glued to other text",
			claims: map[string]interface{}{
				"piped": `{{ .User.CustomClaims "affiliation" | printf "%v" }}`,
			},
			user: dto.UserJWT{}.WithCustomClaims(json.RawMessage(`{"affiliation": ["student", "staff"]}`)),
			// Go's default %v formatting of a []interface{}, not JSON syntax - a known,
			// accepted limitation of gluing an array/object into a larger expression (there's
			// no well-defined "typed" form for that anyway). Scalars are unaffected: %v on a
			// string/number/bool renders identically to before.
			expectedClaims: json.RawMessage(`{"piped": "[student staff]"}`),
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
