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
	"testing"
)

func TestUserHandlerAdminSuite(t *testing.T) {
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

	count, err := s.Storage.GetUserPersister().Count(uuid.Nil, "")
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

	count, err := s.Storage.GetUserPersister().Count(uuid.Nil, "")
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

func (s *userAdminSuite) TestUserHandlerAdmin_List_InvalidPaginationParam() {
	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/users?per_page=invalid", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}
