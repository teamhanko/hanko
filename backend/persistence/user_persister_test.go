package persistence_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestUserPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(userPersisterSuite))
}

type userPersisterSuite struct {
	test.Suite
}

const (
	userPersisterTenantID = "00000000-0000-0000-0000-000000000001"
	userPersisterOrgID    = "cafecafe-0000-0000-0000-000000000001"
	userPersisterRoleID   = "cafecafe-0000-0000-0000-000000000002" // slug "admin"
	userPersisterMemberID = "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"
	userPersisterOtherID  = "38bf5a00-d7ea-40a5-a5de-48722c148925" // no memberships
)

func (s *userPersisterSuite) addMembershipAndBinding() {
	membershipID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetOrganizationMembershipPersister().Create(models.OrganizationMembership{
		ID:             membershipID,
		TenantID:       uuid.FromStringOrNil(userPersisterTenantID),
		UserID:         uuid.FromStringOrNil(userPersisterMemberID),
		OrganizationID: uuid.FromStringOrNil(userPersisterOrgID),
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)

	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       uuid.FromStringOrNil(userPersisterTenantID),
		UserID:         uuid.FromStringOrNil(userPersisterMemberID),
		RoleID:         uuid.FromStringOrNil(userPersisterRoleID),
		OrganizationID: uuid.FromStringOrNil(userPersisterOrgID),
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)
}

func (s *userPersisterSuite) TestGet_IncludesOrganizations() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)
	s.addMembershipAndBinding()

	tenantID := uuid.FromStringOrNil(userPersisterTenantID)

	member, err := s.Storage.GetUserPersister().Get(uuid.FromStringOrNil(userPersisterMemberID), tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(member)
	s.Require().Len(member.Organizations, 1)
	s.Equal(uuid.FromStringOrNil(userPersisterOrgID), member.Organizations[0].OrganizationID)
	s.Require().Len(member.Organizations[0].Roles, 1)
	s.Equal("admin", member.Organizations[0].Roles[0].Slug)

	other, err := s.Storage.GetUserPersister().Get(uuid.FromStringOrNil(userPersisterOtherID), tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(other)
	s.Empty(other.Organizations)
}

func (s *userPersisterSuite) TestList_IncludesOrganizationsPerUser() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)
	s.addMembershipAndBinding()

	tenantID := uuid.FromStringOrNil(userPersisterTenantID)
	userIDs := []uuid.UUID{
		uuid.FromStringOrNil(userPersisterMemberID),
		uuid.FromStringOrNil(userPersisterOtherID),
	}

	users, err := s.Storage.GetUserPersister().List(1, 20, userIDs, "", "", "desc", tenantID)
	s.Require().NoError(err)

	byID := map[uuid.UUID][]models.UserOrganizationRoles{}
	for _, u := range users {
		byID[u.ID] = u.Organizations
	}

	s.Require().Len(byID[uuid.FromStringOrNil(userPersisterMemberID)], 1)
	s.Empty(byID[uuid.FromStringOrNil(userPersisterOtherID)])
}

func (s *userPersisterSuite) TestGetByUsername_IncludesOrganizations() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_admin")
	s.Require().NoError(err)

	membershipID, err := uuid.NewV4()
	s.Require().NoError(err)
	tenantID := uuid.FromStringOrNil(userPersisterTenantID)
	usernameUserID := uuid.FromStringOrNil("6be8e0e7-05e6-4223-807b-06fe1d5e7b75")

	err = s.Storage.GetOrganizationMembershipPersister().Create(models.OrganizationMembership{
		ID:             membershipID,
		TenantID:       tenantID,
		UserID:         usernameUserID,
		OrganizationID: uuid.FromStringOrNil(userPersisterOrgID),
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)

	user, err := s.Storage.GetUserPersister().GetByUsername("foouser", tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(user)
	s.Require().Len(user.Organizations, 1)
	s.Equal(uuid.FromStringOrNil(userPersisterOrgID), user.Organizations[0].OrganizationID)
}
