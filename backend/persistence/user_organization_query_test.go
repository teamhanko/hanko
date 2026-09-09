package persistence_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestListOrganizationsWithRolesByUserIDsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(listOrganizationsWithRolesByUserIDsSuite))
}

type listOrganizationsWithRolesByUserIDsSuite struct {
	test.Suite
}

const role1bID = "55555555-5555-5555-5555-555555555556" // "admin", tenant1

func (s *listOrganizationsWithRolesByUserIDsSuite) createMembership(userID uuid.UUID, organizationID uuid.UUID) {
	membershipID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetOrganizationMembershipPersister().Create(models.OrganizationMembership{
		ID:             membershipID,
		TenantID:       tenant1ID,
		UserID:         userID,
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)
}

func (s *listOrganizationsWithRolesByUserIDsSuite) createBinding(userID uuid.UUID, roleID uuid.UUID, organizationID uuid.UUID) {
	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       tenant1ID,
		UserID:         userID,
		RoleID:         roleID,
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)
}

// A user in two organizations, one with two roles and one with none, must
// come back correctly grouped: two organizations, the right roles under
// each, and an empty (not missing) role list for the one with none.
func (s *listOrganizationsWithRolesByUserIDsSuite) TestGroupsMultipleOrgsAndRoles() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership(user1ID, org1ID)
	s.createMembership(user1ID, org1bID)
	s.createBinding(user1ID, role1ID, org1ID)
	s.createBinding(user1ID, uuid.FromStringOrNil(role1bID), org1ID)

	result, err := s.Storage.GetOrganizationMembershipPersister().ListOrganizationsWithRolesByUserIDs([]uuid.UUID{user1ID}, tenant1ID)
	s.Require().NoError(err)

	orgs := result[user1ID]
	s.Require().Len(orgs, 2)

	byOrgID := map[uuid.UUID]models.UserOrganizationRoles{}
	for _, org := range orgs {
		byOrgID[org.OrganizationID] = org
	}

	s.Require().Contains(byOrgID, org1ID)
	s.Len(byOrgID[org1ID].Roles, 2)

	s.Require().Contains(byOrgID, org1bID)
	s.Empty(byOrgID[org1bID].Roles)
}

// Batching multiple users into one call must not cross-contaminate their
// results - each user must only see their own organizations.
func (s *listOrganizationsWithRolesByUserIDsSuite) TestBatchesMultipleUsersWithoutCrossContamination() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership(user1ID, org1ID)
	s.createBinding(user1ID, role1ID, org1ID)

	s.createMembership(user1bID, org1bID)
	s.createBinding(user1bID, uuid.FromStringOrNil(role1bID), org1bID)

	result, err := s.Storage.GetOrganizationMembershipPersister().
		ListOrganizationsWithRolesByUserIDs([]uuid.UUID{user1ID, user1bID}, tenant1ID)
	s.Require().NoError(err)

	s.Require().Len(result[user1ID], 1)
	s.Equal(org1ID, result[user1ID][0].OrganizationID)

	s.Require().Len(result[user1bID], 1)
	s.Equal(org1bID, result[user1bID][0].OrganizationID)
}

func (s *listOrganizationsWithRolesByUserIDsSuite) TestUserWithNoMembershipsHasNoEntry() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	result, err := s.Storage.GetOrganizationMembershipPersister().
		ListOrganizationsWithRolesByUserIDs([]uuid.UUID{user1ID}, tenant1ID)
	s.Require().NoError(err)
	s.Empty(result[user1ID])
}

func (s *listOrganizationsWithRolesByUserIDsSuite) TestEmptyUserIDsReturnsEmptyMap() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	result, err := s.Storage.GetOrganizationMembershipPersister().
		ListOrganizationsWithRolesByUserIDs(nil, tenant1ID)
	s.Require().NoError(err)
	s.Empty(result)
}
