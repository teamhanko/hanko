package persistence_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestOrganizationPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(organizationPersisterSuite))
}

type organizationPersisterSuite struct {
	test.Suite
}

// Organization name uniqueness is scoped per tenant, not global - the same
// name must be creatable in two different tenants without conflict.
func (s *organizationPersisterSuite) TestCreate_AllowsSameNameAcrossTenants() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	idA, err := uuid.NewV4()
	s.Require().NoError(err)
	idB, err := uuid.NewV4()
	s.Require().NoError(err)

	now := time.Now()
	err = s.Storage.GetOrganizationPersister().Create(models.Organization{
		ID:        idA,
		TenantID:  tenant1ID,
		Name:      "Shared Org Name",
		CreatedAt: now,
		UpdatedAt: now,
	})
	s.Require().NoError(err)

	err = s.Storage.GetOrganizationPersister().Create(models.Organization{
		ID:        idB,
		TenantID:  tenant2ID,
		Name:      "Shared Org Name",
		CreatedAt: now,
		UpdatedAt: now,
	})
	s.Require().NoError(err)
}

// A tenant-scoped Get must not resolve an organization that belongs to a
// different tenant, even though the id itself is valid and exists.
func (s *organizationPersisterSuite) TestGet_DoesNotLeakAcrossTenants() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	// org2ID belongs to tenant2 - looking it up under tenant1 must return nil.
	organization, err := s.Storage.GetOrganizationPersister().Get(org2ID, tenant1ID)
	s.Require().NoError(err)
	s.Nil(organization)
}
