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

// TestPublicIdUniqueness is a migration/schema-level test: (tenant_id, public_id) uniqueness is
// enforced, but the same public_id may be reused across different tenants, and NULL public_id
// values never collide with each other, even within the same tenant.
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

	// Same public_id, different tenants: must succeed - a public_id is unique per tenant, not globally.
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
