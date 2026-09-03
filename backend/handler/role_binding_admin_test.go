package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestRoleBindingHandlerAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(roleBindingAdminSuite))
}

type roleBindingAdminSuite struct {
	test.Suite
}

const (
	roleBindingTestOrgID    = "55555555-6666-0000-0000-000000000001"
	roleBindingTestMemberID = "44444444-5555-0000-0000-000000000001"
	roleBindingTestOtherID  = "44444444-5555-0000-0000-000000000002" // not a member
	roleBindingTestRoleID   = "66666666-7777-0000-0000-000000000001" // "admin", not yet bound
	roleBindingTestSlug     = "member"                               // "member", already bound
)

func (s *roleBindingAdminSuite) TestRoleBindingHandlerAdmin_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		userID             string
		body               string
		expectedStatusCode int
	}{
		{
			name:               "success by id",
			userID:             roleBindingTestMemberID,
			body:               fmt.Sprintf(`{"role": "%s"}`, roleBindingTestRoleID),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "already bound",
			userID:             roleBindingTestMemberID,
			body:               fmt.Sprintf(`{"role": "%s"}`, roleBindingTestSlug),
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "not a member",
			userID:             roleBindingTestOtherID,
			body:               fmt.Sprintf(`{"role": "%s"}`, roleBindingTestSlug),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "unrecognized role",
			userID:             roleBindingTestMemberID,
			body:               `{"role": "does-not-exist"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "missing role",
			userID:             roleBindingTestMemberID,
			body:               `{}`,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/role_binding_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/organizations/%s/users/%s/roles", roleBindingTestOrgID, currentTest.userID), strings.NewReader(currentTest.body))
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

func (s *roleBindingAdminSuite) TestRoleBindingHandlerAdmin_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/role_binding_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/users/%s/roles", roleBindingTestOrgID, roleBindingTestMemberID), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var got []map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	s.Require().NoError(err)
	s.Len(got, 1)
	s.Equal(roleBindingTestSlug, got[0]["slug"])
}

func (s *roleBindingAdminSuite) TestRoleBindingHandlerAdmin_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		userID             string
		roleRef            string
		expectedStatusCode int
	}{
		{
			name:               "success by slug",
			userID:             roleBindingTestMemberID,
			roleRef:            roleBindingTestSlug,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "unrecognized role",
			userID:             roleBindingTestMemberID,
			roleRef:            "does-not-exist",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "no such binding",
			userID:             roleBindingTestOtherID,
			roleRef:            roleBindingTestSlug,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/role_binding_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/organizations/%s/users/%s/roles/%s", roleBindingTestOrgID, currentTest.userID, currentTest.roleRef), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}
