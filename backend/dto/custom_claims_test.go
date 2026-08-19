package dto

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func TestCustomClaimsJWTFromUserModel_UnwrapsSourceEnvelope(t *testing.T) {
	model := &models.UserCustomClaims{
		UserID: uuid.Must(uuid.NewV4()),
		Claims: json.RawMessage(`{
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
	assert.Nil(t, CustomClaimsJWTFromUserModel(&models.UserCustomClaims{Claims: json.RawMessage(`{}`)}))
}

func TestCustomClaimsJWTFromUserModel_InvalidJSONOmitted(t *testing.T) {
	model := &models.UserCustomClaims{
		UserID: uuid.Must(uuid.NewV4()),
		Claims: json.RawMessage(`not valid json`),
	}

	assert.Nil(t, CustomClaimsJWTFromUserModel(model))
}
