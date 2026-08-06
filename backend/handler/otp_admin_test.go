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
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestOtpAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(otpAdminSuite))
}

type otpAdminSuite struct {
	test.Suite
}

func (s *otpAdminSuite) TestOtpAdminHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/otp")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userId             string
		expectedStatusCode int
	}{
		{
			name:               "should return otp credential",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should fail if user has no otp credential",
			userId:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty userID",
			userId:             "",
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/otp", currentTest.userId), nil)

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if http.StatusOK == rec.Code {
				var otpCredential *admin.OTPDto
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &otpCredential))
				s.NotNil(otpCredential)
			} else {
				s.Require().Equal(currentTest.expectedStatusCode, rec.Code)
			}
		})
	}
}

func (s *otpAdminSuite) TestOtpAdminHandler_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/otp")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userId             string
		expectedStatusCode int
	}{
		{
			name:               "should delete the otp credential",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should fail if user has no otp credential",
			userId:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty userID",
			userId:             "",
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
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/otp", currentTest.userId), nil)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Require().Equal(currentTest.expectedStatusCode, rec.Code)

			if http.StatusNoContent == rec.Code {
				cred, err := s.Storage.GetPasswordCredentialPersister().GetByUserID(uuid.FromStringOrNil(currentTest.userId), uuid.FromStringOrNil(config.DefaultTenantID))
				s.Require().NoError(err)
				s.Require().Nil(cred)
			}
		})
	}
}

// TestOtpAdminHandler_Get_UsesPublicIdNotInternalId proves :user_id resolves via public_id and
// not the real internal id, for a handler that was fixed but had no dedicated regression test of
// its own.
func (s *otpAdminSuite) TestOtpAdminHandler_Get_UsesPublicIdNotInternalId() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/otp_admin_public_id")
	s.Require().NoError(err)

	cfg := test.DefaultConfig
	err = cfg.PostProcess()
	s.Require().NoError(err)

	e := NewAdminRouter(&cfg, s.Storage, nil)

	internalID := "266aa04a-7070-4a11-8a11-707070707070"
	publicID := "377bb04a-8080-4b22-8b22-808080808080"

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/otp", internalID), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code, "the real internal id must not resolve the user or their otp secret")

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/otp", publicID), nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code, "the public_id must resolve the user and their otp secret")
}
