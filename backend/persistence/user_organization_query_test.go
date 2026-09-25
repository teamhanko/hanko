package persistence_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestUserOrganizationsGroupingSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(userOrganizationsGroupingSuite))
}

type userOrganizationsGroupingSuite struct {
	test.Suite
}

const role1bID = "55555555-5555-5555-5555-555555555556" // "admin", tenant1

func (s *userOrganizationsGroupingSuite) createMembership(userID uuid.UUID, organizationID uuid.UUID) {
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

func (s *userOrganizationsGroupingSuite) createBinding(userID uuid.UUID, roleID uuid.UUID, organizationID uuid.UUID) {
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
func (s *userOrganizationsGroupingSuite) TestGet_GroupsMultipleOrgsAndRoles() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership(user1ID, org1ID)
	s.createMembership(user1ID, org1bID)
	s.createBinding(user1ID, role1ID, org1ID)
	s.createBinding(user1ID, uuid.FromStringOrNil(role1bID), org1ID)

	user, err := s.Storage.GetUserPersister().Get(user1ID, tenant1ID)
	s.Require().NoError(err)
	s.Require().NotNil(user)

	orgs := user.Organizations
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

// Listing multiple users at once must not cross-contaminate their
// organizations - each user must only see their own.
func (s *userOrganizationsGroupingSuite) TestList_BatchesMultipleUsersWithoutCrossContamination() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership(user1ID, org1ID)
	s.createBinding(user1ID, role1ID, org1ID)

	s.createMembership(user1bID, org1bID)
	s.createBinding(user1bID, uuid.FromStringOrNil(role1bID), org1bID)

	users, err := s.Storage.GetUserPersister().
		List(1, 20, []uuid.UUID{user1ID, user1bID}, "", "", "desc", tenant1ID)
	s.Require().NoError(err)

	byID := map[uuid.UUID][]models.UserOrganizationRoles{}
	for _, u := range users {
		byID[u.ID] = u.Organizations
	}

	s.Require().Len(byID[user1ID], 1)
	s.Equal(org1ID, byID[user1ID][0].OrganizationID)

	s.Require().Len(byID[user1bID], 1)
	s.Equal(org1bID, byID[user1bID][0].OrganizationID)
}

func (s *userOrganizationsGroupingSuite) TestGet_UserWithNoMembershipsHasEmptyOrganizations() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	user, err := s.Storage.GetUserPersister().Get(user1ID, tenant1ID)
	s.Require().NoError(err)
	s.Require().NotNil(user)
	s.Empty(user.Organizations)
}
