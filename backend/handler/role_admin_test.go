package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestRoleHandlerAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(roleAdminSuite))
}

type roleAdminSuite struct {
	test.Suite
}

func (s *roleAdminSuite) TestRoleHandlerAdmin_Create() {
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
			body:               `{"slug": "billing-manager", "name": "Billing Manager"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "missing slug",
			body:               `{"name": "No Slug"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "missing name",
			body:               `{"slug": "no-name"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with already existing slug",
			body:               `{"slug": "admin", "name": "Duplicate Admin"}`,
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "uuid-shaped slug",
			body:               `{"slug": "123e4567-e89b-12d3-a456-426614174000", "name": "Sneaky"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/role_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(currentTest.body))
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

func (s *roleAdminSuite) TestRoleHandlerAdmin_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/role_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("2", rec.Header().Get("X-Total-Count"))

	var got []map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	s.Require().NoError(err)
	s.Len(got, 2)
}

func (s *roleAdminSuite) TestRoleHandlerAdmin_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		roleID             string
		expectedStatusCode int
	}{
		{
			name:               "success",
			roleID:             "dddddddd-0000-0000-0000-000000000001",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "not found",
			roleID:             "00000000-0000-0000-0000-000000000099",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "invalid id",
			roleID:             "not-a-uuid",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/role_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/roles/%s", currentTest.roleID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}

func (s *roleAdminSuite) TestRoleHandlerAdmin_Patch() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		roleID             string
		body               string
		expectedStatusCode int
		verify             func(s *roleAdminSuite, rec *httptest.ResponseRecorder)
	}{
		{
			name:               "success renames name only",
			roleID:             "dddddddd-0000-0000-0000-000000000001",
			body:               `{"name": "Renamed Admin"}`,
			expectedStatusCode: http.StatusOK,
			verify: func(s *roleAdminSuite, rec *httptest.ResponseRecorder) {
				var got map[string]any
				err := json.Unmarshal(rec.Body.Bytes(), &got)
				s.Require().NoError(err)
				s.Equal("Renamed Admin", got["name"])
				// slug is not a field on PatchRoleRequest, so it can't change via PATCH
				s.Equal("admin", got["slug"])
			},
		},
		{
			name:               "empty name",
			roleID:             "dddddddd-0000-0000-0000-000000000001",
			body:               `{"name": ""}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "not found",
			roleID:             "00000000-0000-0000-0000-000000000099",
			body:               `{"name": "Doesn't Matter"}`,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/role_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/roles/%s", currentTest.roleID), strings.NewReader(currentTest.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)
			if currentTest.verify != nil {
				currentTest.verify(s, rec)
			}

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}

func (s *roleAdminSuite) TestRoleHandlerAdmin_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/role_admin")
	s.Require().NoError(err)

	// dddddddd-...-001 ("admin") has no bindings in the fixture set.
	req := httptest.NewRequest(http.MethodDelete, "/roles/dddddddd-0000-0000-0000-000000000001", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)

	count, err := s.Storage.GetRolePersister().Count(uuid.FromStringOrNil(config.DefaultTenantID))
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s *roleAdminSuite) TestRoleHandlerAdmin_Delete_NotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/role_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodDelete, "/roles/00000000-0000-0000-0000-000000000099", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}

// A role belonging to a different tenant must never be reachable through
// another tenant's admin API - Get, Patch, and Delete all resolve through
// the same tenant-scoped persister lookup, so this covers that shared
// code path.
func (s *roleAdminSuite) TestRoleHandlerAdmin_Get_DoesNotLeakAcrossTenants() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	cfg := test.DefaultConfig
	cfg.MultiTenancy.Enabled = true
	err := cfg.PostProcess()
	s.Require().NoError(err)
	e := NewAdminRouter(&cfg, s.Storage, nil)
	defer e.Close()

	err = s.LoadFixtures("../test/fixtures/role_admin")
	s.Require().NoError(err)

	// dddddddd-...-003 belongs to tenant 2 - fetching it under tenant 1's
	// path must 404, not leak the role across tenants.
	req := httptest.NewRequest(http.MethodGet, "/00000000-0000-0000-0000-000000000001/roles/dddddddd-0000-0000-0000-000000000003", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *roleAdminSuite) TestRoleHandlerAdmin_Delete_CascadesToRoleBindings() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/role_admin")
	s.Require().NoError(err)

	// dddddddd-...-002 ("member") has a role_binding in the fixture set.
	req := httptest.NewRequest(http.MethodDelete, "/roles/dddddddd-0000-0000-0000-000000000002", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)

	count, err := s.Storage.GetRolePersister().Count(uuid.FromStringOrNil(config.DefaultTenantID))
	s.Require().NoError(err)
	s.Equal(1, count)

	binding, err := s.Storage.GetRoleBindingPersister().Get(
		uuid.FromStringOrNil("bbbbbbbb-0000-0000-0000-000000000001"),
		uuid.FromStringOrNil("dddddddd-0000-0000-0000-000000000002"),
		uuid.FromStringOrNil("cccccccc-0000-0000-0000-000000000001"),
		uuid.FromStringOrNil(config.DefaultTenantID),
	)
	s.Require().NoError(err)
	s.Nil(binding)
}

// Deleting a role must cascade to its role_bindings rows whether the
// deployment runs in single- or multi-tenant mode.
func (s *roleAdminSuite) TestRoleHandlerAdmin_Delete_CascadesToRoleBindings_MultiTenantRouting() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	cfg := test.DefaultConfig
	cfg.MultiTenancy.Enabled = true
	err := cfg.PostProcess()
	s.Require().NoError(err)
	e := NewAdminRouter(&cfg, s.Storage, nil)
	defer e.Close()

	err = s.LoadFixtures("../test/fixtures/role_admin")
	s.Require().NoError(err)

	// dddddddd-...-002 ("member") has a role_binding in the fixture set.
	req := httptest.NewRequest(http.MethodDelete, "/00000000-0000-0000-0000-000000000001/roles/dddddddd-0000-0000-0000-000000000002", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)

	// The binding is gone - deleting the role cascaded to it.
	binding, err := s.Storage.GetRoleBindingPersister().Get(
		uuid.FromStringOrNil("bbbbbbbb-0000-0000-0000-000000000001"),
		uuid.FromStringOrNil("dddddddd-0000-0000-0000-000000000002"),
		uuid.FromStringOrNil("cccccccc-0000-0000-0000-000000000001"),
		uuid.FromStringOrNil(config.DefaultTenantID),
	)
	s.Require().NoError(err)
	s.Nil(binding)
}
