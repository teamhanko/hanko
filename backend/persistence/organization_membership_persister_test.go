package persistence_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestOrganizationMembershipPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(organizationMembershipPersisterSuite))
}

type organizationMembershipPersisterSuite struct {
	test.Suite
}

var (
	tenant1ID = uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")
	tenant2ID = uuid.FromStringOrNil("00000000-0000-0000-0000-000000000002")
	user1ID   = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
	user2ID   = uuid.FromStringOrNil("22222222-2222-2222-2222-222222222222")
	org1ID    = uuid.FromStringOrNil("33333333-3333-3333-3333-333333333333")
	org2ID    = uuid.FromStringOrNil("44444444-4444-4444-4444-444444444444")
	role1ID   = uuid.FromStringOrNil("55555555-5555-5555-5555-555555555555")
	role2ID   = uuid.FromStringOrNil("66666666-6666-6666-6666-666666666666")
	// A second user and organization, both in tenant1 - for scenarios where
	// the same role is bound to different users across different
	// organizations within one tenant.
	user1bID = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111114")
	org1bID  = uuid.FromStringOrNil("33333333-3333-3333-3333-333333333336")
)

// This is the core tenant-isolation guarantee from the design: a membership
// row must not be creatable if the user and organization it links belong to
// different tenants, even though both rows individually exist.
func (s *organizationMembershipPersisterSuite) TestCreate_RejectsCrossTenantUser() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	id, err := uuid.NewV4()
	s.Require().NoError(err)

	// user1 belongs to tenant1, org2 belongs to tenant2 - this must fail.
	membership := models.OrganizationMembership{
		ID:             id,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		OrganizationID: org2ID,
		CreatedAt:      time.Now(),
	}

	err = s.Storage.GetOrganizationMembershipPersister().Create(membership)
	s.Require().Error(err)
}

func (s *organizationMembershipPersisterSuite) TestCreate_RejectsCrossTenantOrganization() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	id, err := uuid.NewV4()
	s.Require().NoError(err)

	// org1 belongs to tenant1, user2 belongs to tenant2 - this must fail.
	membership := models.OrganizationMembership{
		ID:             id,
		TenantID:       tenant1ID,
		UserID:         user2ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	}

	err = s.Storage.GetOrganizationMembershipPersister().Create(membership)
	s.Require().Error(err)
}

func (s *organizationMembershipPersisterSuite) TestOrganizationDelete_CascadesToMemberships() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	membershipID, err := uuid.NewV4()
	s.Require().NoError(err)

	membership := models.OrganizationMembership{
		ID:             membershipID,
		TenantID:       tenant1ID,
		UserID:         user1ID,
		OrganizationID: org1ID,
		CreatedAt:      time.Now(),
	}
	err = s.Storage.GetOrganizationMembershipPersister().Create(membership)
	s.Require().NoError(err)

	organization, err := s.Storage.GetOrganizationPersister().Get(org1ID, tenant1ID)
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	err = s.Storage.GetOrganizationPersister().Delete(*organization)
	s.Require().NoError(err)

	remaining, err := s.Storage.GetOrganizationMembershipPersister().Get(user1ID, org1ID, tenant1ID)
	s.Require().NoError(err)
	s.Nil(remaining)
}
