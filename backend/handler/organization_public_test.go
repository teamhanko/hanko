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
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestOrganizationPublicHandlerSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(organizationPublicSuite))
}

type organizationPublicSuite struct {
	test.Suite
}

const (
	rolePublicTestTenantID  = "00000000-0000-0000-0000-000000000001"
	rolePublicTestMemberID  = "99999999-1111-0000-0000-000000000001" // member of org, holds "admin"
	rolePublicTestOtherID   = "99999999-1111-0000-0000-000000000002" // not a member of anything
	rolePublicTestOrgID     = "99999999-2222-0000-0000-000000000001"
	rolePublicTestAdminID   = "99999999-3333-0000-0000-000000000001" // slug "admin", held by member
	rolePublicTestBillingID = "99999999-3333-0000-0000-000000000002" // slug "billing-manager", held by no one
)

func (s *organizationPublicSuite) newRouter() (http.Handler, func() error) {
	cfg := test.DefaultConfig
	err := cfg.PostProcess()
	s.Require().NoError(err)
	err = local_db.SyncSecretKeys(&cfg, s.Storage)
	s.Require().NoError(err)

	e := NewPublicRouter(&cfg, s.Storage, nil, nil)
	return e, e.Close
}

func (s *organizationPublicSuite) TestOrganizationPublicHandler_CheckRole() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name            string
		userID          string
		body            string
		expectedHasRole bool
	}{
		{
			name:            "holds role, referenced by slug",
			userID:          rolePublicTestMemberID,
			body:            `{"organization_id": "` + rolePublicTestOrgID + `", "roles": ["admin"]}`,
			expectedHasRole: true,
		},
		{
			name:            "holds role, referenced by id",
			userID:          rolePublicTestMemberID,
			body:            `{"organization_id": "` + rolePublicTestOrgID + `", "roles": ["` + rolePublicTestAdminID + `"]}`,
			expectedHasRole: true,
		},
		{
			name:            "does not hold role",
			userID:          rolePublicTestMemberID,
			body:            `{"organization_id": "` + rolePublicTestOrgID + `", "roles": ["billing-manager"]}`,
			expectedHasRole: false,
		},
		{
			name:            "OR semantics - holds one of several",
			userID:          rolePublicTestMemberID,
			body:            `{"organization_id": "` + rolePublicTestOrgID + `", "roles": ["billing-manager", "admin"]}`,
			expectedHasRole: true,
		},
		{
			name:            "not a member of the organization",
			userID:          rolePublicTestOtherID,
			body:            `{"organization_id": "` + rolePublicTestOrgID + `", "roles": ["admin"]}`,
			expectedHasRole: false,
		},
		{
			name:            "unrecognized organization",
			userID:          rolePublicTestMemberID,
			body:            `{"organization_id": "00000000-0000-0000-0000-000000000099", "roles": ["admin"]}`,
			expectedHasRole: false,
		},
		{
			name:            "unrecognized role",
			userID:          rolePublicTestMemberID,
			body:            `{"organization_id": "` + rolePublicTestOrgID + `", "roles": ["does-not-exist"]}`,
			expectedHasRole: false,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e, closeRouter := s.newRouter()

			err := s.LoadFixtures("../test/fixtures/organization_public")
			s.Require().NoError(err)

			cookie, err := generateSessionCookie(s.Storage, uuid.FromStringOrNil(currentTest.userID), uuid.FromStringOrNil(rolePublicTestTenantID))
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/organizations/roles/check", strings.NewReader(currentTest.body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(http.StatusOK, rec.Code)

			var got map[string]any
			err = json.Unmarshal(rec.Body.Bytes(), &got)
			s.Require().NoError(err)
			s.Equal(currentTest.expectedHasRole, got["has_role"])

			s.Require().NoError(closeRouter())
			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}

// An organization belonging to a different tenant must resolve the same
// way as an organization that doesn't exist at all: has_role false, never
// an error - so a real organization id from tenant 2 must not leak into
// tenant 1's role check (nor be treated any differently from a garbage
// id, which the CheckRole test above already covers). Exercised under
// real multi-tenant routing, not just single-tenant's fixed default
// tenant.
func (s *organizationPublicSuite) TestOrganizationPublicHandler_CheckRole_DoesNotLeakAcrossTenants() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	err := s.LoadFixtures("../test/fixtures/organization_public")
	s.Require().NoError(err)

	err = generateSigningKeyForTenant(s.Storage, uuid.FromStringOrNil(rolePublicTestTenantID))
	s.Require().NoError(err)

	cfg := test.DefaultConfig
	cfg.MultiTenancy.Enabled = true
	err = cfg.PostProcess()
	s.Require().NoError(err)

	e := NewPublicRouter(&cfg, s.Storage, nil, nil)
	defer e.Close()

	cookie, err := generateSessionCookie(s.Storage, uuid.FromStringOrNil(rolePublicTestMemberID), uuid.FromStringOrNil(rolePublicTestTenantID))
	s.Require().NoError(err)

	// 99999999-2222-...-0002 is a real organization, but it belongs to
	// tenant 2 - the session above, and the request path, are tenant 1's.
	body := `{"organization_id": "99999999-2222-0000-0000-000000000002", "roles": ["admin"]}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s/organizations/roles/check", rolePublicTestTenantID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var got map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	s.Require().NoError(err)
	s.Equal(false, got["has_role"])
}

func (s *organizationPublicSuite) TestOrganizationPublicHandler_CheckRole_Unauthenticated() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e, closeRouter := s.newRouter()
	defer closeRouter()

	err := s.LoadFixtures("../test/fixtures/organization_public")
	s.Require().NoError(err)

	body := `{"organization_id": "` + rolePublicTestOrgID + `", "roles": ["admin"]}`
	req := httptest.NewRequest(http.MethodPost, "/organizations/roles/check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *organizationPublicSuite) TestOrganizationPublicHandler_CheckRole_InvalidBody() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name string
		body string
	}{
		{name: "missing organization_id", body: `{"roles": ["admin"]}`},
		{name: "missing roles", body: `{"organization_id": "` + rolePublicTestOrgID + `"}`},
		{name: "empty roles", body: `{"organization_id": "` + rolePublicTestOrgID + `", "roles": []}`},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e, closeRouter := s.newRouter()

			err := s.LoadFixtures("../test/fixtures/organization_public")
			s.Require().NoError(err)

			cookie, err := generateSessionCookie(s.Storage, uuid.FromStringOrNil(rolePublicTestMemberID), uuid.FromStringOrNil(rolePublicTestTenantID))
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/organizations/roles/check", strings.NewReader(currentTest.body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(http.StatusBadRequest, rec.Code)

			s.Require().NoError(closeRouter())
			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}
