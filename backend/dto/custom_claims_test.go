package dto

import (
	"encoding/json"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func TestCustomClaimsFromUserModel_UnwrapsSourceEnvelope(t *testing.T) {
	model := &models.UserCustomClaims{
		UserID: uuid.Must(uuid.NewV4()),
		Claims: nulls.NewString(`{
			"matriculation_number": {"value": "12345", "source": "saml:uni-a"},
			"is_staff": {"value": true, "source": "admin"}
		}`),
	}

	claims := CustomClaimsFromUserModel(model)
	user := UserJWT{}.WithCustomClaims(claims)

	assert.NotNil(t, claims)
	assert.Equal(t, "12345", user.CustomClaims("matriculation_number"))
	assert.Equal(t, true, user.CustomClaims("is_staff"))
	assert.NotContains(t, string(claims), "source", "internal source tag must never reach the JWT")
	assert.NotContains(t, string(claims), "saml:uni-a")
}

func TestCustomClaimsFromUserModel_NilModel(t *testing.T) {
	assert.Nil(t, CustomClaimsFromUserModel(nil))
}

func TestCustomClaimsFromUserModel_EmptyClaims(t *testing.T) {
	assert.Nil(t, CustomClaimsFromUserModel(&models.UserCustomClaims{}))
	assert.Nil(t, CustomClaimsFromUserModel(&models.UserCustomClaims{Claims: nulls.NewString(`{}`)}))
}

func TestCustomClaimsFromUserModel_InvalidJSONOmitted(t *testing.T) {
	model := &models.UserCustomClaims{
		UserID: uuid.Must(uuid.NewV4()),
		Claims: nulls.NewString(`not valid json`),
	}

	assert.Nil(t, CustomClaimsFromUserModel(model))
}

// A user with no resolved custom claims yet (no UserCustomClaims row, or an empty one) leaves
// UserJWT.customClaims nil by design - see dto/user.go. A JWT template referencing
// `.User.CustomClaims "some_claim"` must not panic in that case, nor must a nil *UserJWT.
func TestUserJWT_CustomClaims_NoData(t *testing.T) {
	var user *UserJWT
	assert.Equal(t, "", user.CustomClaims("matriculation_number"))
	assert.Equal(t, "", user.CustomClaims())

	empty := UserJWT{}
	assert.Equal(t, "", empty.CustomClaims("matriculation_number"))
	assert.Equal(t, "", empty.CustomClaims())
}

// CustomClaims returns each claim with its real JSON type intact (float64/bool/string/
// []interface{}), not stringified - see its doc comment for why that matters beyond just a
// bare claim template's own output (session.ProcessJWTTemplate's `{{if}}`/`{{and}}`/`{{not}}`/
// `eq` support depends on it too).
func TestUserJWT_CustomClaims_TypedValues(t *testing.T) {
	user := UserJWT{}.WithCustomClaims(json.RawMessage(`{
		"matriculation_number": "12345",
		"age": 29,
		"is_staff": true,
		"affiliation": ["student", "staff"]
	}`))

	assert.Equal(t, "12345", user.CustomClaims("matriculation_number"))
	assert.Equal(t, float64(29), user.CustomClaims("age"))
	assert.Equal(t, true, user.CustomClaims("is_staff"))
	assert.Equal(t, []interface{}{"student", "staff"}, user.CustomClaims("affiliation"))
	assert.Equal(t, "", user.CustomClaims("not_a_real_claim"))

	whole, ok := user.CustomClaims().(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(29), whole["age"])

	var nilUser *UserJWT
	assert.Equal(t, "", nilUser.CustomClaims("age"))
	assert.Equal(t, "", nilUser.CustomClaims())
}
