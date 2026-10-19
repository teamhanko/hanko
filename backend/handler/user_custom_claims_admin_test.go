package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestUserCustomClaimsAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(userCustomClaimsAdminSuite))
}

type userCustomClaimsAdminSuite struct {
	test.Suite
}

// customClaimsTestConfig is a local copy of test.DefaultConfig with custom_claims.definitions
// declared, so it doesn't mutate (or depend on mutating) the shared global other tests rely
// on. Single-tenant mode ignores the tenant DB row's config column entirely - definitions
// have to come from here.
func customClaimsTestConfig() config.Config {
	cfg := test.DefaultConfig
	cfg.TenantConfig.CustomClaims = config.CustomClaims{
		Definitions: config.CustomClaimDefinitions{
			"matriculation_number": {Name: "matriculation_number", Type: config.CustomClaimTypeString},
			"is_staff":             {Name: "is_staff", Type: config.CustomClaimTypeBoolean},
			"affiliation":          {Name: "affiliation", Type: config.CustomClaimTypeStringList},
		},
	}
	return cfg
}

func (s *userCustomClaimsAdminSuite) TestGetCustomClaims() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		userId             string
		expectedStatusCode int
		expectedClaims     admin.CustomClaims
	}{
		{
			name:               "should return custom claims for user with custom claims",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusOK,
			expectedClaims:     admin.CustomClaims{"matriculation_number": "12345", "is_staff": true},
		},
		{
			name:               "should return no content for user without custom claims",
			userId:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on non existing user",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			err := s.LoadFixtures("../test/fixtures/custom_claims")
			s.Require().NoError(err)

			cfg := customClaimsTestConfig()
			e := NewAdminRouter(&cfg, s.Storage, nil)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/custom_claims", currentTest.userId), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Require().Equal(currentTest.expectedStatusCode, rec.Code)

			if currentTest.expectedStatusCode == http.StatusOK {
				var response admin.CustomClaims
				s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &response))
				s.Require().Equal(currentTest.expectedClaims, response)
			}
		})
	}
}

// TestGetCustomClaims_DoesNotLeakAcrossTenants guards against the handler resolving a user by
// public_id without also scoping to the requesting tenant - a user with the given public_id in
// a different tenant must resolve as not found, not as that other tenant's user.
func (s *userCustomClaimsAdminSuite) TestGetCustomClaims_DoesNotLeakAcrossTenants() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/custom_claims")
	s.Require().NoError(err)

	cfg := customClaimsTestConfig()
	cfg.MultiTenancy.Enabled = true
	err = cfg.PostProcess()
	s.Require().NoError(err)
	e := NewAdminRouter(&cfg, s.Storage, nil)

	tenant1ID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")
	tenant2UserID := uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/00000000-0000-0000-0000-000000000001/users/%s/custom_claims", tenant2UserID),
		nil,
	)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)

	// Query for row existence directly (not via GetUserCustomClaimsPersister().Get, which
	// auto-creates a row on miss and would otherwise mask the very bug this test checks for).
	exists, err := s.Storage.GetConnection().
		Where("user_id = ? AND tenant_id = ?", tenant2UserID, tenant1ID).
		Exists(&models.UserCustomClaims{})
	s.Require().NoError(err)
	s.False(exists, "no custom claims row should have been auto-created for the cross-tenant user")
}
