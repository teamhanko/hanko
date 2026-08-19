package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
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

func (s *userCustomClaimsAdminSuite) TestPatchCustomClaims() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	const existingUser = "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	tests := []struct {
		name               string
		userId             string
		patch              string
		expectedStatusCode int
		expectedClaims     admin.CustomClaims
	}{
		{
			name:               "should set a declared string claim",
			userId:             existingUser,
			patch:              `{"matriculation_number":"99999"}`,
			expectedStatusCode: http.StatusOK,
			expectedClaims:     admin.CustomClaims{"matriculation_number": "99999", "is_staff": true},
		},
		{
			name:               "should merge in a new declared claim",
			userId:             existingUser,
			patch:              `{"affiliation":["student","staff"]}`,
			expectedStatusCode: http.StatusOK,
			expectedClaims: admin.CustomClaims{
				"matriculation_number": "12345",
				"is_staff":             true,
				"affiliation":          []any{"student", "staff"},
			},
		},
		{
			name:               "should clear a single claim via null",
			userId:             existingUser,
			patch:              `{"is_staff":null}`,
			expectedStatusCode: http.StatusOK,
			expectedClaims:     admin.CustomClaims{"matriculation_number": "12345"},
		},
		{
			name:               "should clear every claim via a top-level null body",
			userId:             existingUser,
			patch:              `null`,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should reject an undeclared claim",
			userId:             existingUser,
			patch:              `{"not_a_real_claim":"value"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should reject a type mismatch",
			userId:             existingUser,
			patch:              `{"is_staff":"true"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should reject a string_list containing a non-string element",
			userId:             existingUser,
			patch:              `{"affiliation":["student",1]}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			patch:              `{"matriculation_number":"1"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on non existing user",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			patch:              `{"matriculation_number":"1"}`,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			err := s.LoadFixtures("../test/fixtures/custom_claims")
			s.Require().NoError(err)

			cfg := customClaimsTestConfig()
			e := NewAdminRouter(&cfg, s.Storage, nil)

			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/users/%s/custom_claims", currentTest.userId),
				strings.NewReader(currentTest.patch),
			)
			req.Header.Set("Content-Type", "application/json")
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
