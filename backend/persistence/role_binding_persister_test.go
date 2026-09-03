package persistence_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestRoleBindingPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(roleBindingPersisterSuite))
}

type roleBindingPersisterSuite struct {
	test.Suite
}

func (s *roleBindingPersisterSuite) createMembership() {
	membershipID, err := uuid.NewV4()
	s.Require().NoError(err)

	err = s.Storage.GetOrganizationMembershipPersister().Create(models.OrganizationMembership{
		ID:             membershipID,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)
}

// A role binding must not be creatable without a matching organization
// membership already existing - this is what the doc calls "creating a role
// binding requires a matching organization_memberships row to already exist",
// backstopped here at the DB level by the composite FK to
// organization_memberships(user_id, organization_id).
func (s *roleBindingPersisterSuite) TestCreate_RequiresExistingMembership() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)

	// No membership was created for user1/org1 - this must fail.
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		RoleID:         role1ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	})
	s.Require().Error(err)
}

func (s *roleBindingPersisterSuite) TestCreate_RejectsCrossTenantRole() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership()

	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)

	// user1/org1 are both tenant1, but role2 belongs to tenant2 - this must fail.
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		RoleID:         role2ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	})
	s.Require().Error(err)
}

// Deleting a role that is still bound to a user must fail (RESTRICT), rather
// than silently revoking it from everyone holding it - this is a deliberate
// decision, distinct from how organization/user deletion cascade.
func (s *roleBindingPersisterSuite) TestRoleDelete_RestrictedWhenBindingsExist() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership()

	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		RoleID:         role1ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)

	role, err := s.Storage.GetRolePersister().Get(role1ID, tenant1ID)
	s.Require().NoError(err)
	s.Require().NotNil(role)

	err = s.Storage.GetRolePersister().Delete(*role)
	s.Require().Error(err)
}

// This is the FK that makes "removing a user from an organization cascades to
// that user's role_bindings for that organization" actually true at the DB
// level - the least obvious FK in the whole design, and the fastest possible
// signal if it's ever missing or dropped in a future migration.
func (s *roleBindingPersisterSuite) TestMembershipDelete_CascadesToRoleBindings() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership()

	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		RoleID:         role1ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)

	membership, err := s.Storage.GetOrganizationMembershipPersister().Get(user1ID, org1ID, tenant1ID)
	s.Require().NoError(err)
	s.Require().NotNil(membership)

	err = s.Storage.GetOrganizationMembershipPersister().Delete(*membership)
	s.Require().NoError(err)

	remaining, err := s.Storage.GetRoleBindingPersister().Get(user1ID, role1ID, org1ID, tenant1ID)
	s.Require().NoError(err)
	s.Nil(remaining)
}

func (s *roleBindingPersisterSuite) TestOrganizationDelete_CascadesToRoleBindings() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	s.createMembership()

	bindingID, err := uuid.NewV4()
	s.Require().NoError(err)
	err = s.Storage.GetRoleBindingPersister().Create(models.RoleBinding{
		ID:             bindingID,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		RoleID:         role1ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	})
	s.Require().NoError(err)

	organization, err := s.Storage.GetOrganizationPersister().Get(org1ID, tenant1ID)
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	err = s.Storage.GetOrganizationPersister().Delete(*organization)
	s.Require().NoError(err)

	remaining, err := s.Storage.GetRoleBindingPersister().Get(user1ID, role1ID, org1ID, tenant1ID)
	s.Require().NoError(err)
	s.Nil(remaining)
}
