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

func TestOrganizationHandlerAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(organizationAdminSuite))
}

type organizationAdminSuite struct {
	test.Suite
}

func (s *organizationAdminSuite) TestOrganizationHandlerAdmin_Create() {
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
			body:               `{"name": "New Org"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "missing name",
			body:               `{}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "with already existing name",
			body:               `{"name": "Acme Corp"}`,
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/organization_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/organizations", strings.NewReader(currentTest.body))
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

func (s *organizationAdminSuite) TestOrganizationHandlerAdmin_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/organization_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("2", rec.Header().Get("X-Total-Count"))

	var got []map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	s.Require().NoError(err)
	s.Len(got, 2)
}

func (s *organizationAdminSuite) TestOrganizationHandlerAdmin_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		organizationID     string
		expectedStatusCode int
	}{
		{
			name:               "success",
			organizationID:     "aaaaaaaa-0000-0000-0000-000000000001",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "not found",
			organizationID:     "00000000-0000-0000-0000-000000000099",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "invalid id",
			organizationID:     "not-a-uuid",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/organization_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s", currentTest.organizationID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}

func (s *organizationAdminSuite) TestOrganizationHandlerAdmin_Patch() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		organizationID     string
		body               string
		expectedStatusCode int
		verify             func(s *organizationAdminSuite, rec *httptest.ResponseRecorder)
	}{
		{
			name:               "success",
			organizationID:     "aaaaaaaa-0000-0000-0000-000000000001",
			body:               `{"name": "Renamed Org"}`,
			expectedStatusCode: http.StatusOK,
			verify: func(s *organizationAdminSuite, rec *httptest.ResponseRecorder) {
				var got map[string]any
				err := json.Unmarshal(rec.Body.Bytes(), &got)
				s.Require().NoError(err)
				s.Equal("Renamed Org", got["name"])
			},
		},
		{
			name:               "empty name",
			organizationID:     "aaaaaaaa-0000-0000-0000-000000000001",
			body:               `{"name": ""}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "not found",
			organizationID:     "00000000-0000-0000-0000-000000000099",
			body:               `{"name": "Doesn't Matter"}`,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/organization_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/organizations/%s", currentTest.organizationID), strings.NewReader(currentTest.body))
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

func (s *organizationAdminSuite) TestOrganizationHandlerAdmin_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/organization_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/aaaaaaaa-0000-0000-0000-000000000001", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)

	count, err := s.Storage.GetOrganizationPersister().Count(uuid.FromStringOrNil(config.DefaultTenantID))
	s.Require().NoError(err)
	s.Equal(1, count)
}

// An organization belonging to a different tenant must never be
// reachable through another tenant's admin API - Get, Patch, and Delete
// all resolve through the same tenant-scoped persister lookup, so this
// covers that shared code path.
func (s *organizationAdminSuite) TestOrganizationHandlerAdmin_Get_DoesNotLeakAcrossTenants() {
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

	err = s.LoadFixtures("../test/fixtures/organization_admin")
	s.Require().NoError(err)

	// aaaaaaaa-...-0003 belongs to tenant 2 - fetching it under tenant 1's
	// path must 404, not leak the organization across tenants.
	req := httptest.NewRequest(http.MethodGet, "/00000000-0000-0000-0000-000000000001/organizations/aaaaaaaa-0000-0000-0000-000000000003", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *organizationAdminSuite) TestOrganizationHandlerAdmin_Delete_NotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/organization_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/00000000-0000-0000-0000-000000000099", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}
