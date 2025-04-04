package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGetClaimsFromToken(t *testing.T) {
	subject := uuid.Must(uuid.NewV4())
	sessionID := uuid.Must(uuid.NewV4())
	now := time.Now()
	expiration := now.Add(1 * time.Hour)

	tests := []struct {
		name          string
		token         jwt.Token
		expected      *Claims
		expectedError string
	}{
		{
			name: "valid token with all claims",
			token: func() jwt.Token {
				token, _ := jwt.NewBuilder().
					Subject(subject.String()).
					IssuedAt(now).
					Audience([]string{"test-audience"}).
					Issuer("test-issuer").
					Expiration(expiration).
					Claim("session_id", sessionID.String()).
					Claim("email", map[string]interface{}{
						"address":     "test@example.com",
						"is_verified": true,
						"is_primary":  true,
					}).
					Claim("username", "testuser").
					Claim("custom", "value").
					Build()
				return token
			}(),
			expected: &Claims{
				Subject:   subject,
				SessionID: sessionID,
				IssuedAt:  &now,
				Audience:  []string{"test-audience"},
				Issuer:    stringPtr("test-issuer"),
				Email: &EmailJWT{
					Address:    "test@example.com",
					IsVerified: true,
					IsPrimary:  true,
				},
				Username:   stringPtr("testuser"),
				Expiration: expiration,
				CustomClaims: map[string]interface{}{
					"custom": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := GetClaimsFromToken(tt.token)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, claims)

			// Compare the claims
			if tt.expected != nil {
				assert.Equal(t, tt.expected.Subject, claims.Subject)
				assert.Equal(t, tt.expected.SessionID, claims.SessionID)
				assert.Equal(t, tt.expected.Audience, claims.Audience)
				assert.Equal(t, tt.expected.Issuer, claims.Issuer)
				assert.Equal(t, tt.expected.Username, claims.Username)
				assert.Equal(t, tt.expected.CustomClaims, claims.CustomClaims)

				if tt.expected.Email != nil {
					assert.Equal(t, tt.expected.Email.Address, claims.Email.Address)
					assert.Equal(t, tt.expected.Email.IsVerified, claims.Email.IsVerified)
					assert.Equal(t, tt.expected.Email.IsPrimary, claims.Email.IsPrimary)
				}
			}
		})
	}
}

func TestClaims_MarshalJSON(t *testing.T) {
	subject := uuid.Must(uuid.NewV4())
	sessionID := uuid.Must(uuid.NewV4())
	now := time.Now().Truncate(time.Second)
	expiration := now.Add(1 * time.Hour)
	username := "testuser"
	issuer := "test-issuer"

	tests := []struct {
		name     string
		claims   Claims
		expected map[string]interface{}
	}{
		{
			name: "all fields populated",
			claims: Claims{
				Subject:   subject,
				SessionID: sessionID,
				IssuedAt:  &now,
				Audience:  []string{"test-audience"},
				Issuer:    &issuer,
				Email: &EmailJWT{
					Address:    "test@example.com",
					IsVerified: true,
					IsPrimary:  true,
				},
				Username:   &username,
				Expiration: expiration,
				CustomClaims: map[string]interface{}{
					"custom": "value",
				},
			},
			expected: map[string]interface{}{
				"subject":    subject.String(),
				"session_id": sessionID.String(),
				"issued_at":  now,
				"audience":   []interface{}{"test-audience"},
				"issuer":     issuer,
				"email": map[string]interface{}{
					"address":     "test@example.com",
					"is_verified": true,
					"is_primary":  true,
				},
				"username":   username,
				"expiration": expiration,
				"custom":     "value",
			},
		},
		{
			name: "minimal fields",
			claims: Claims{
				Subject:    subject,
				SessionID:  sessionID,
				Expiration: expiration,
			},
			expected: map[string]interface{}{
				"subject":    subject.String(),
				"session_id": sessionID.String(),
				"expiration": expiration,
			},
		},
		{
			name: "with custom claims only",
			claims: Claims{
				Subject:    subject,
				SessionID:  sessionID,
				Expiration: expiration,
				CustomClaims: map[string]interface{}{
					"custom1": "value1",
					"custom2": "value2",
				},
			},
			expected: map[string]interface{}{
				"subject":    subject.String(),
				"session_id": sessionID.String(),
				"expiration": expiration,
				"custom1":    "value1",
				"custom2":    "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the claims to JSON
			jsonData, err := json.Marshal(tt.claims)
			assert.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Unmarshal the JSON back to a map for comparison
			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			// Compare the expected and actual results
			for key, expectedValue := range tt.expected {
				actualValue := result[key]
				switch v := expectedValue.(type) {
				case time.Time:
					// For time values, compare the string representation after truncating to seconds
					expectedTime := v.Truncate(time.Second).UTC()
					actualTime, err := time.Parse(time.RFC3339, actualValue.(string))
					assert.NoError(t, err)
					actualTime = actualTime.Truncate(time.Second).UTC()
					assert.Equal(t, expectedTime, actualTime, "time mismatch for key: %s", key)
				case *time.Time:
					// For pointer to time values, compare the string representation after truncating to seconds
					expectedTime := v.Truncate(time.Second).UTC()
					actualTime, err := time.Parse(time.RFC3339, actualValue.(string))
					assert.NoError(t, err)
					actualTime = actualTime.Truncate(time.Second).UTC()
					assert.Equal(t, expectedTime, actualTime, "time mismatch for key: %s", key)
				case uuid.UUID:
					// For UUID values, compare the string representation
					assert.Equal(t, v.String(), actualValue, "UUID mismatch for key: %s", key)
				default:
					assert.Equal(t, expectedValue, actualValue, "mismatch for key: %s", key)
				}
			}
		})
	}
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
