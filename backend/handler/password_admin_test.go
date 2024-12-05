package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPasswordAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(passwordAdminSuite))
}

type passwordAdminSuite struct {
	test.Suite
}

func (s *passwordAdminSuite) TestPasswordAdminHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/password")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userId             string
		expectedStatusCode int
	}{
		{
			name:   "should return password credential",
			userId: "38bf5a00-d7ea-40a5-a5de-48722c148925",
		},
		{
			name:               "should fail if user has no password",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/password", currentTest.userId), nil)

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if http.StatusOK == rec.Code {
				var passwordCredential *admin.PasswordCredential
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &passwordCredential))
				s.NotNil(passwordCredential)
			} else {
				s.Require().Equal(currentTest.expectedStatusCode, rec.Code)
			}
		})
	}
}

func (s *passwordAdminSuite) TestPasswordAdminHandler_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/password")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userId             string
		password           string
		expectedStatusCode int
	}{
		{
			name:     "should create a new password credential",
			userId:   "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			password: "superSecure",
		},
		{
			name:               "should fail if user already has a password",
			userId:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			password:           "superSecure",
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			password:           "superSecure",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty userID",
			userId:             "",
			password:           "superSecure",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on non existing user",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			password:           "superSecure",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on empty password",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			password:           "",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			testDto := admin.CreateOrUpdatePasswordCredentialRequestDto{
				GetPasswordCredentialRequestDto: admin.GetPasswordCredentialRequestDto{
					UserID: currentTest.userId,
				},
				Password: currentTest.password,
			}

			testJson, err := json.Marshal(testDto)
			s.Require().NoError(err)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/password", currentTest.userId), bytes.NewReader(testJson))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if http.StatusOK == rec.Code {
				var passwordCredential *admin.PasswordCredential
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &passwordCredential))
				s.NotNil(passwordCredential)

				cred, err := s.Storage.GetPasswordCredentialPersister().GetByUserID(uuid.FromStringOrNil(currentTest.userId))
				s.Require().NoError(err)
				s.Require().NotNil(cred)
			} else {
				s.Require().Equal(currentTest.expectedStatusCode, rec.Code)
			}
		})
	}
}

func (s *passwordAdminSuite) TestPasswordAdminHandler_Update() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/password")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userId             string
		oldHashedPassword  string
		password           string
		expectedStatusCode int
	}{
		{
			name:              "should update a password credential",
			userId:            "38bf5a00-d7ea-40a5-a5de-48722c148925",
			oldHashedPassword: "$2a$12$Cf7k.dG6pznTUJ5u2u1pgu6I4VXH5.9O0NZsDk8TwWwyBkZovYVli",
			password:          "superSecure",
		},
		{
			name:               "should fail if user already has no password",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			password:           "superSecure",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non uuid userID",
			userId:             "customUserId",
			password:           "superSecure",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty userID",
			userId:             "",
			password:           "superSecure",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on non existing user",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			password:           "superSecure",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on empty password",
			userId:             "30f41697-b413-43cc-8cca-d55298683607",
			password:           "",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			testDto := admin.CreateOrUpdatePasswordCredentialRequestDto{
				GetPasswordCredentialRequestDto: admin.GetPasswordCredentialRequestDto{
					UserID: currentTest.userId,
				},
				Password: currentTest.password,
			}

			testJson, err := json.Marshal(testDto)
			s.Require().NoError(err)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s/password", currentTest.userId), bytes.NewReader(testJson))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if http.StatusOK == rec.Code {
				var passwordCredential *admin.PasswordCredential
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &passwordCredential))
				s.NotNil(passwordCredential)

				cred, err := s.Storage.GetPasswordCredentialPersister().GetByUserID(uuid.FromStringOrNil(currentTest.userId))
				s.Require().NoError(err)
				s.NotEqual(currentTest.oldHashedPassword, cred.Password)
			} else {
				s.Require().Equal(currentTest.expectedStatusCode, rec.Code)
			}
		})
	}
}

func (s *passwordAdminSuite) TestPasswordAdminHandler_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/password")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userId             string
		expectedStatusCode int
	}{
		{
			name:               "should delete a password credential",
			userId:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should fail if user already has no password",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
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
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/password", currentTest.userId), nil)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Require().Equal(currentTest.expectedStatusCode, rec.Code)

			if http.StatusNoContent == rec.Code {
				cred, err := s.Storage.GetPasswordCredentialPersister().GetByUserID(uuid.FromStringOrNil(currentTest.userId))
				s.Require().NoError(err)
				s.Require().Nil(cred)
			}
		})
	}
}
