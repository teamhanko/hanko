package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v2/dto/admin"
	"github.com/teamhanko/hanko/backend/v2/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(sessionAdminSuite))
}

type sessionAdminSuite struct {
	test.Suite
}

func (s *sessionAdminSuite) TestSessionAdminHandler_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userID             string
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "should return a list of sessions with multiple entries",
			userID:             "ec4ef049-5b88-4321-a173-21b0eff06a04",
			expectedStatusCode: http.StatusOK,
			expectedCount:      2,
		},
		{
			name:               "should return a list of sessions with one entry",
			userID:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "should return an empty list",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			expectedStatusCode: http.StatusOK,
			expectedCount:      0,
		},
		{
			name:               "should fail on non uuid userID",
			userID:             "customUserId",
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
		},
		{
			name:               "should fail on empty userID",
			userID:             "",
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      0,
		},
		{
			name:               "should fail on non existing user",
			userID:             "30f41697-b413-43cc-8cca-d55298683607",
			expectedStatusCode: http.StatusNotFound,
			expectedCount:      0,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/sessions", currentTest.userID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)
			if http.StatusOK == rec.Code {
				var sessions []admin.ListSessionsRequestDto
				err = json.Unmarshal(rec.Body.Bytes(), &sessions)
				s.Require().NoError(err)

				s.Equal(currentTest.expectedCount, len(sessions))
			}
		})
	}
}

func (s *sessionAdminSuite) TestSessionAdminHandler_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userID             string
		sessionID          string
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "should delete session for user with multiple sessions",
			userID:             "ec4ef049-5b88-4321-a173-21b0eff06a04",
			sessionID:          "d8d6dc27-fcf9-4a5c-bb50-a7a03067d936",
			expectedCount:      1,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should delete session for user with one session",
			userID:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			sessionID:          "108f3789-a795-43bd-a58f-ac8e80a213cd",
			expectedCount:      0,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should fail if session is not found",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			sessionID:          "649c95d7-9840-4e6d-be00-6c6b93c9e885",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail if session is not associated to the user",
			userID:             "38bf5a00-d7ea-40a5-a5de-48722c148925",
			sessionID:          "74ba812a-923a-43e4-8020-9535dcadc0a8",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non existing user",
			userID:             "30f41697-b413-43cc-8cca-d55298683607",
			sessionID:          "6e405e60-f70c-4b8a-b0d5-8ba05dd3e793",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on empty userID",
			userID:             "",
			sessionID:          "6e405e60-f70c-4b8a-b0d5-8ba05dd3e793",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty sessionID",
			userID:             "46626836-f2db-4ec0-8752-858b544cbc78",
			sessionID:          "",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail on non uuid userID",
			userID:             "customUserId",
			sessionID:          "d8d6dc27-fcf9-4a5c-bb50-a7a03067d936",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/sessions/%s", currentTest.userID, currentTest.sessionID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)
			if http.StatusNoContent == rec.Code {
				credentials, err := s.Storage.GetSessionPersister().ListActive(uuid.FromStringOrNil(currentTest.userID))
				s.Require().NoError(err)
				s.Equal(currentTest.expectedCount, len(credentials))
			}
		})
	}
}
