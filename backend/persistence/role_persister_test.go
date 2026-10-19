package persistence_test

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestRolePersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(rolePersisterSuite))
}

type rolePersisterSuite struct {
	test.Suite
}

// Role slug uniqueness is scoped per tenant, not global - the same slug
// must be creatable in two different tenants without conflict.
func (s *rolePersisterSuite) TestCreate_AllowsSameSlugAcrossTenants() {
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
	err = s.Storage.GetRolePersister().Create(models.Role{
		ID:        idA,
		TenantID:  tenant1ID,
		Slug:      "shared-slug",
		Name:      "Shared",
		CreatedAt: now,
		UpdatedAt: now,
	})
	s.Require().NoError(err)

	err = s.Storage.GetRolePersister().Create(models.Role{
		ID:        idB,
		TenantID:  tenant2ID,
		Slug:      "shared-slug",
		Name:      "Shared",
		CreatedAt: now,
		UpdatedAt: now,
	})
	s.Require().NoError(err)
}

// GetByIDOrSlug must not resolve a role belonging to a different tenant,
// whether referenced by id or by slug - this is what makes the resolver
// safe to use from tenant-scoped callers (role bindings, user creation,
// the public role-check endpoint). role1ID ("member", tenant1) and
// role2ID ("member", tenant2) deliberately share a slug, so resolving
// "member" under tenant1 must return role1, never role2.
func (s *rolePersisterSuite) TestGetByIDOrSlug_DoesNotLeakAcrossTenants() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/organization_persister")
	s.Require().NoError(err)

	p := s.Storage.GetRolePersister()

	byID, err := p.GetByIDOrSlug(role2ID.String(), tenant1ID)
	s.Require().NoError(err)
	s.Nil(byID)

	bySlug, err := p.GetByIDOrSlug("member", tenant1ID)
	s.Require().NoError(err)
	s.Require().NotNil(bySlug)
	s.Equal(role1ID, bySlug.ID)
}
