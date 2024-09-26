package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserHandlerAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(userAdminSuite))
}

type userAdminSuite struct {
	test.Suite
}

func (s *userAdminSuite) TestUserHandlerAdmin_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", "38bf5a00-d7ea-40a5-a5de-48722c148925"), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)

	count, err := s.Storage.GetUserPersister().Count([]uuid.UUID{}, "", "")
	s.Require().NoError(err)
	s.Equal(2, count)
}

func (s *userAdminSuite) TestUserHandlerAdmin_Delete_UnknownUserId() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", "1e5dcc5c-8570-43cb-ba8b-caa88bbfc7ac"), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)

	count, err := s.Storage.GetUserPersister().Count([]uuid.UUID{}, "", "")
	s.Require().NoError(err)
	s.Equal(3, count)
}

func (s *userAdminSuite) TestUserHandlerAdmin_Delete_InvalidUserId() {
	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", "invalidId"), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *userAdminSuite) TestUserHandlerAdmin_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("3", rec.Header().Get("X-Total-Count"))
	s.Equal("<http://example.com/users?page=1&per_page=20>; rel=\"first\"", rec.Header().Get("Link"))
}

func (s *userAdminSuite) TestUserHandlerAdmin_List_Pagination() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/users?page=1&per_page=1", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		s.Equal("3", rec.Header().Get("X-Total-Count"))

		var got []models.User
		err = json.Unmarshal(rec.Body.Bytes(), &got)
		s.Require().NoError(err)

		s.Equal(1, len(got))
		s.Equal("<http://example.com/users?page=3&per_page=1>; rel=\"last\",<http://example.com/users?page=2&per_page=1>; rel=\"next\"", rec.Header().Get("Link"))
	}
}

func (s *userAdminSuite) TestUserHandlerAdmin_List_NoUsers() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		s.Equal("0", rec.Header().Get("X-Total-Count"))

		var got []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &got)
		s.Require().NoError(err)

		s.Equal(0, len(got))
		s.Equal("<http://example.com/users?page=1&per_page=20>; rel=\"first\"", rec.Header().Get("Link"))
	}
}

func (s *userAdminSuite) TestUserHandlerAdmin_List_MultipleUserIDs() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/users?user_id=b5dd5267-b462-48be-b70d-bcd6f1bbe7a5,e0282f3f-b211-4f0e-b777-6fabc69287c9", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		s.Equal("2", rec.Header().Get("X-Total-Count"))

		var got []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &got)
		s.Require().NoError(err)

		s.Equal(2, len(got))
		s.Equal("<http://example.com/users?page=1&per_page=20&user_id=b5dd5267-b462-48be-b70d-bcd6f1bbe7a5%2Ce0282f3f-b211-4f0e-b777-6fabc69287c9>; rel=\"first\"", rec.Header().Get("Link"))
	}
}

func (s *userAdminSuite) TestUserHandlerAdmin_List_InvalidPaginationParam() {
	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/users?per_page=invalid", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *userAdminSuite) TestUserHandlerAdmin_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		body               string
		expectedStatusCode int
	}{
		{
			name:               "success",
			body:               `{"emails": [{"address": "test@test.com", "is_primary": true}]}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "success with user id",
			body:               `{"id": "98a46ea2-51f6-4e30-bd29-8272de77c8c8", "emails": [{"address": "test@test.com", "is_primary": true}]}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "success with multiple emails",
			body:               `{"emails": [{"address": "test@test.com", "is_primary": true}, {"address": "test2@test.com"}]}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "success with created_at",
			body:               `{"emails": [{"address": "test@test.com", "is_primary": true}], "created_at": "2023-06-07T13:42:49.369489Z"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "with already existing user id",
			body:               `{"id": "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5", "emails": [{"address": "test@test.com", "is_primary": true}]}`,
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "with non uuid v4 id",
			body:               `{"id": "customId", "emails": [{"address": "test@test.com", "is_primary": true}]}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with no emails",
			body:               `{"id": "98a46ea2-51f6-4e30-bd29-8272de77c8c8", "emails": []}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with missing emails",
			body:               `{"id": "98a46ea2-51f6-4e30-bd29-8272de77c8c8"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with no primary email",
			body:               `{"emails": [{"address": "test@test.com"}]}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with multiple primary emails",
			body:               `{"emails": [{"address": "test@test.com", "is_primary": true"}, {"address": "test2@test.com", "is_primary": true"}]}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with non unique emails",
			body:               `{"emails": [{"address": "test@test.com", "is_primary": true"}, {"address": "test@test.com"}]}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with already existing email",
			body:               `{"emails": [{"address": "john.doe@example.com", "is_primary": true}]}`,
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/user_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(currentTest.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}
