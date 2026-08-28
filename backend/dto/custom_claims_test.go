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
	assert.Equal(t, "true", user.CustomClaims("is_staff"))
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

// CustomClaimsValue is CustomClaims's type-preserving counterpart, used only by
// session.parseClaimTemplateValue's bare-accessor fast path (see its doc comment) - it should
// return the value's real JSON type, not the stringified form CustomClaims returns.
func TestUserJWT_CustomClaimsValue(t *testing.T) {
	user := UserJWT{}.WithCustomClaims(json.RawMessage(`{
		"matriculation_number": "12345",
		"age": 29,
		"is_staff": true,
		"affiliation": ["student", "staff"]
	}`))

	assert.Equal(t, "12345", user.CustomClaimsValue("matriculation_number"))
	assert.Equal(t, float64(29), user.CustomClaimsValue("age"))
	assert.Equal(t, true, user.CustomClaimsValue("is_staff"))
	assert.Equal(t, []interface{}{"student", "staff"}, user.CustomClaimsValue("affiliation"))
	assert.Equal(t, "", user.CustomClaimsValue("not_a_real_claim"))

	whole, ok := user.CustomClaimsValue().(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(29), whole["age"])

	var nilUser *UserJWT
	assert.Equal(t, "", nilUser.CustomClaimsValue("age"))
	assert.Equal(t, "", nilUser.CustomClaimsValue())
}
