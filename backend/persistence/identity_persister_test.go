package persistence_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestIdentityPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(identityPersisterSuite))
}

type identityPersisterSuite struct {
	test.Suite
}

const (
	identityPersisterTenantID       = "00000000-0000-0000-0000-000000000001"
	identityPersisterIdentityID     = "9d2e4f60-7b81-4c92-a3d4-e5f6a7b8c901"
	identityPersisterSamlProviderID = "7c9e6679-7425-40de-944b-e07fc1f90ae8"
	identityPersisterProviderID     = "saml_" + identityPersisterSamlProviderID
	identityPersisterProviderUserID = "jane-doe-name-id"
)

func (s *identityPersisterSuite) TestGet_AttachesSamlProvider() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/saml_provider_lookup")
	s.Require().NoError(err)

	identity, err := s.Storage.GetIdentityPersister().Get(identityPersisterProviderUserID, identityPersisterProviderID, uuid.FromStringOrNil(identityPersisterTenantID))
	s.Require().NoError(err)
	s.Require().NotNil(identity)
	s.Require().NotNil(identity.SamlIdentity)
	s.Require().NotNil(identity.SamlIdentity.SamlProvider)
	s.Equal(identityPersisterSamlProviderID, identity.SamlIdentity.SamlProvider.ID.String())
}

func (s *identityPersisterSuite) TestGetByID_AttachesSamlProvider() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/saml_provider_lookup")
	s.Require().NoError(err)

	identity, err := s.Storage.GetIdentityPersister().GetByID(uuid.FromStringOrNil(identityPersisterIdentityID), uuid.FromStringOrNil(identityPersisterTenantID))
	s.Require().NoError(err)
	s.Require().NotNil(identity)
	s.Require().NotNil(identity.SamlIdentity)
	s.Require().NotNil(identity.SamlIdentity.SamlProvider)
	s.Equal(identityPersisterSamlProviderID, identity.SamlIdentity.SamlProvider.ID.String())
}
