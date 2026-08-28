package dto

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func TestCustomClaimsJWTFromUserModel_UnwrapsSourceEnvelope(t *testing.T) {
	model := &models.UserCustomClaims{
		UserID: uuid.Must(uuid.NewV4()),
		Claims: nulls.NewString(`{
			"matriculation_number": {"value": "12345", "source": "saml:uni-a"},
			"is_staff": {"value": true, "source": "admin"}
		}`),
	}

	jwt := CustomClaimsJWTFromUserModel(model)

	assert.NotNil(t, jwt)
	assert.Equal(t, "12345", jwt.Get("matriculation_number"))
	assert.Equal(t, "true", jwt.Get("is_staff"))
	assert.NotContains(t, jwt.String(), "source", "internal source tag must never reach the JWT")
	assert.NotContains(t, jwt.String(), "saml:uni-a")
}

func TestCustomClaimsJWTFromUserModel_NilModel(t *testing.T) {
	assert.Nil(t, CustomClaimsJWTFromUserModel(nil))
}

func TestCustomClaimsJWTFromUserModel_EmptyClaims(t *testing.T) {
	assert.Nil(t, CustomClaimsJWTFromUserModel(&models.UserCustomClaims{}))
	assert.Nil(t, CustomClaimsJWTFromUserModel(&models.UserCustomClaims{Claims: nulls.NewString(`{}`)}))
}

func TestCustomClaimsJWTFromUserModel_InvalidJSONOmitted(t *testing.T) {
	model := &models.UserCustomClaims{
		UserID: uuid.Must(uuid.NewV4()),
		Claims: nulls.NewString(`not valid json`),
	}

	assert.Nil(t, CustomClaimsJWTFromUserModel(model))
}

// A user with no resolved custom claims yet (no UserCustomClaims row, or an empty one) leaves
// UserJWT.CustomClaims nil by design - see dto/user.go. A JWT template referencing
// `.User.CustomClaims.Get "some_claim"` must not panic in that case.
func TestCustomClaimsJWT_Get_NilReceiver(t *testing.T) {
	var jwt *CustomClaimsJWT

	assert.Equal(t, "", jwt.Get("matriculation_number"))
	assert.Equal(t, "", jwt.Get())
}
