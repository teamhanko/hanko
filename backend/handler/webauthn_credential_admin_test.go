package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebauthnCredentialAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(webauthnCredentialAdminSuite))
}

type webauthnCredentialAdminSuite struct {
	test.Suite
}

func (s *webauthnCredentialAdminSuite) TestWebauthnCredentialAdminHandler_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/webauthn")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userID             string
		expectedCount      int
		expectedStatusCode int
	}{
		{
			name:               "should return webauthn credentials for user with multiple credentials",
			userID:             "ec4ef049-5b88-4321-a173-21b0eff06a04",
			expectedCount:      2,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should return webauthn credentials for user with one credentials",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			expectedCount:      1,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should return webauthn credentials for user with no credentials",
			userID:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			expectedCount:      0,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should fail on non uuid userID",
			userID:             "customUserId",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty userID",
			userID:             "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on non existing user",
			userID:             "30f41697-b413-43cc-8cca-d55298683607",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/webauthn_credentials", currentTest.userID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)
			if http.StatusOK == rec.Code {
				var credentials []dto.WebauthnCredentialResponse
				err = json.Unmarshal(rec.Body.Bytes(), &credentials)
				s.Require().NoError(err)

				s.Equal(len(credentials), currentTest.expectedCount)
			}
		})
	}
}

func (s *webauthnCredentialAdminSuite) TestWebauthnCredentialAdminHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/webauthn")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userID             string
		credentialID       string
		expectedStatusCode int
	}{
		{
			name:               "should return webauthn credential",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should fail if credential is not found",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			credentialID:       "notSoRandomCredentialID",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail if credential is not associated to the user",
			userID:             "ec4ef049-5b88-4321-a173-21b0eff06a04",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non existing user",
			userID:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a6",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on empty userID",
			userID:             "",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty credentialID",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			credentialID:       "",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non uuid userID",
			userID:             "customUserId",
			credentialID:       "46626836-f2db-4ec0-8752-858b544cbc78",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/webauthn_credentials/%s", currentTest.userID, currentTest.credentialID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)
			if http.StatusOK == rec.Code {
				var credential dto.WebauthnCredentialResponse
				err = json.Unmarshal(rec.Body.Bytes(), &credential)
				s.Require().NoError(err)
				s.Equal(currentTest.credentialID, credential.ID)
			}
		})
	}
}

func (s *webauthnCredentialAdminSuite) TestWebauthnCredentialAdminHandler_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/webauthn")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userID             string
		credentialID       string
		expectedCount      int
		expectedStatusCode int
	}{
		{
			name:               "should delete webauthn credential for user with multiple credentials",
			userID:             "ec4ef049-5b88-4321-a173-21b0eff06a04",
			credentialID:       "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
			expectedCount:      1,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should delete webauthn credential for user with one credential",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedCount:      0,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should fail if credential is not found",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			credentialID:       "notSoRandomCredentialID",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail if credential is not associated to the user",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			credentialID:       "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjK",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non existing user",
			userID:             "30f41697-b413-43cc-8cca-d55298683607",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on empty userID",
			userID:             "",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty credentialID",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			credentialID:       "",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non uuid userID",
			userID:             "customUserId",
			credentialID:       "4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/webauthn_credentials/%s", currentTest.userID, currentTest.credentialID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)
			if http.StatusNoContent == rec.Code {
				credentials, err := s.Storage.GetWebauthnCredentialPersister().GetFromUser(uuid.FromStringOrNil(currentTest.userID))
				s.Require().NoError(err)
				s.Equal(currentTest.expectedCount, len(credentials))
			}
		})
	}
}
