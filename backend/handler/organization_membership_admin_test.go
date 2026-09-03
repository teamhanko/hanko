package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestOrganizationMembershipAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(organizationMembershipAdminSuite))
}

type organizationMembershipAdminSuite struct {
	test.Suite
}

const (
	membershipTestOrgID       = "22222222-3333-0000-0000-000000000001"
	membershipTestMemberID    = "11111111-2222-0000-0000-000000000001"
	membershipTestNonMemberID = "11111111-2222-0000-0000-000000000002"
)

func (s *organizationMembershipAdminSuite) TestOrganizationHandlerAdmin_AddMember() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		orgID              string
		userID             string
		expectedStatusCode int
	}{
		{
			name:               "success",
			orgID:              membershipTestOrgID,
			userID:             membershipTestNonMemberID,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "already a member",
			orgID:              membershipTestOrgID,
			userID:             membershipTestMemberID,
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:               "unknown organization",
			orgID:              "00000000-0000-0000-0000-000000000099",
			userID:             membershipTestNonMemberID,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "unknown user",
			orgID:              membershipTestOrgID,
			userID:             "00000000-0000-0000-0000-000000000099",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/organization_membership_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/organizations/%s/users/%s", currentTest.orgID, currentTest.userID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}

func (s *organizationMembershipAdminSuite) TestOrganizationHandlerAdmin_RemoveMember() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name               string
		orgID              string
		userID             string
		expectedStatusCode int
	}{
		{
			name:               "success",
			orgID:              membershipTestOrgID,
			userID:             membershipTestMemberID,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "not a member",
			orgID:              membershipTestOrgID,
			userID:             membershipTestNonMemberID,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "unknown organization",
			orgID:              "00000000-0000-0000-0000-000000000099",
			userID:             membershipTestMemberID,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())
			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			err := s.LoadFixtures("../test/fixtures/organization_membership_admin")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/organizations/%s/users/%s", currentTest.orgID, currentTest.userID), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}

func (s *organizationMembershipAdminSuite) TestOrganizationHandlerAdmin_RemoveMember_CascadesToRoleBindings() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/organization_membership_admin")
	s.Require().NoError(err)

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	orgID := uuid.FromStringOrNil(membershipTestOrgID)
	memberID := uuid.FromStringOrNil(membershipTestMemberID)

	now := time.Now()

	roleID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetRolePersister().Create(models.Role{
		ID:        roleID,
		TenantID:  tenantID,
		Slug:      "temp-role",
		Name:      "Temp Role",
		CreatedAt: now,
		UpdatedAt: now,
	})
	s.Require().NoError(err)

	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       tenantID,
		UserID:         memberID,
		RoleID:         roleID,
		OrganizationID: orgID,
		CreatedAt:      now,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/organizations/%s/users/%s", membershipTestOrgID, membershipTestMemberID), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)

	remaining, err := s.Storage.GetRoleBindingPersister().Get(memberID, roleID, orgID, tenantID)
	s.Require().NoError(err)
	s.Nil(remaining)
}

func (s *organizationMembershipAdminSuite) TestOrganizationHandlerAdmin_ListMembers() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/organization_membership_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/users", membershipTestOrgID), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("1", rec.Header().Get("X-Total-Count"))

	var got []map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	s.Require().NoError(err)
	s.Len(got, 1)
	s.Equal(membershipTestMemberID, got[0]["user_id"])
}

func (s *organizationMembershipAdminSuite) TestOrganizationHandlerAdmin_ListMembers_UnknownOrganization() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.Require().NoError(s.Storage.MigrateUp())
	defer func() { s.Require().NoError(s.Storage.MigrateDown(-1)) }()

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)
	defer e.Close()

	err := s.LoadFixtures("../test/fixtures/organization_membership_admin")
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/organizations/00000000-0000-0000-0000-000000000099/users", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}
