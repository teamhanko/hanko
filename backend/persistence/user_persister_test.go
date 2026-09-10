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

// TestPublicIdUniqueness is the migration/schema-level test the plan calls for: (tenant_id,
// public_id) uniqueness is enforced, but the same public_id may be reused across different
// tenants - the actual core deliverable of this whole change - and NULL public_id values never
// collide with each other, even within the same tenant.
func (s *userPersisterSuite) TestPublicIdUniqueness() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tenantAID, err := uuid.NewV4()
	s.Require().NoError(err)
	tenantBID, err := uuid.NewV4()
	s.Require().NoError(err)

	now := time.Now()
	tenantPersister := s.Storage.GetTenantPersister()
	s.Require().NoError(tenantPersister.Create(models.Tenant{ID: tenantAID, CreatedAt: now, UpdatedAt: now}))
	s.Require().NoError(tenantPersister.Create(models.Tenant{ID: tenantBID, CreatedAt: now, UpdatedAt: now}))

	userPersister := s.Storage.GetUserPersister()

	sharedPublicID, err := uuid.NewV4()
	s.Require().NoError(err)

	newUser := func(tenantID uuid.UUID, publicID *uuid.UUID) models.User {
		id, err := uuid.NewV4()
		s.Require().NoError(err)
		return models.User{
			ID:        id,
			PublicID:  publicID,
			TenantID:  tenantID,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	// Same public_id, different tenants: must succeed - this is the actual point of the change.
	s.NoError(userPersister.Create(newUser(tenantAID, &sharedPublicID)))
	s.NoError(userPersister.Create(newUser(tenantBID, &sharedPublicID)))

	// Same public_id, same tenant again: must fail on the (tenant_id, public_id) unique index.
	s.Error(userPersister.Create(newUser(tenantAID, &sharedPublicID)))

	// NULL public_id never collides, even repeatedly within the same tenant.
	s.NoError(userPersister.Create(newUser(tenantAID, nil)))
	s.NoError(userPersister.Create(newUser(tenantAID, nil)))
}

// TestGetByPublicID confirms the persister method itself: resolves a user by their tenant-scoped
// public_id, returns nil (not an error) when no match exists, and never matches across tenants.
func (s *userPersisterSuite) TestGetByPublicID() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tenantAID, err := uuid.NewV4()
	s.Require().NoError(err)
	tenantBID, err := uuid.NewV4()
	s.Require().NoError(err)

	now := time.Now()
	tenantPersister := s.Storage.GetTenantPersister()
	s.Require().NoError(tenantPersister.Create(models.Tenant{ID: tenantAID, CreatedAt: now, UpdatedAt: now}))
	s.Require().NoError(tenantPersister.Create(models.Tenant{ID: tenantBID, CreatedAt: now, UpdatedAt: now}))

	userPersister := s.Storage.GetUserPersister()

	internalID, err := uuid.NewV4()
	s.Require().NoError(err)
	publicID, err := uuid.NewV4()
	s.Require().NoError(err)

	s.Require().NoError(userPersister.Create(models.User{
		ID:        internalID,
		PublicID:  &publicID,
		TenantID:  tenantAID,
		CreatedAt: now,
		UpdatedAt: now,
	}))

	found, err := userPersister.GetByPublicID(publicID, tenantAID)
	s.NoError(err)
	s.Require().NotNil(found)
	s.Equal(internalID, found.ID)

	notFound, err := userPersister.GetByPublicID(publicID, tenantBID)
	s.NoError(err)
	s.Nil(notFound, "a public_id must never resolve across tenants")

	notFoundEither, err := userPersister.GetByPublicID(internalID, tenantAID)
	s.NoError(err)
	s.Nil(notFoundEither, "the real internal id must not work as a public_id lookup key")
}
