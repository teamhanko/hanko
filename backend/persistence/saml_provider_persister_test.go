package persistence_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestSamlProviderPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlProviderPersisterSuite))
}

type samlProviderPersisterSuite struct {
	test.Suite
}

const samlProviderPersisterTenantID = "00000000-0000-0000-0000-000000000001"

func (s *samlProviderPersisterSuite) TestGet_ReturnsNilWhenNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	provider, err := s.Storage.GetSamlProviderPersister().Get(uuid.FromStringOrNil(samlProviderPersisterTenantID), uuid.Must(uuid.NewV4()))
	s.Require().NoError(err)
	s.Nil(provider)
}

func (s *samlProviderPersisterSuite) TestGetByEntityID_ReturnsNilWhenNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	provider, err := s.Storage.GetSamlProviderPersister().GetByEntityID(uuid.FromStringOrNil(samlProviderPersisterTenantID), "https://idp.unknown.example.com/metadata")
	s.Require().NoError(err)
	s.Nil(provider)
}

func (s *samlProviderPersisterSuite) TestGetByDomain_ReturnsNilWhenNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	provider, err := s.Storage.GetSamlProviderPersister().GetByDomain(uuid.FromStringOrNil(samlProviderPersisterTenantID), "unknown.example.com")
	s.Require().NoError(err)
	s.Nil(provider)
}

func (s *samlProviderPersisterSuite) TestGetEnabledByDomain_ReturnsNilWhenNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	provider, err := s.Storage.GetSamlProviderPersister().GetEnabledByDomain(uuid.FromStringOrNil(samlProviderPersisterTenantID), "unknown.example.com")
	s.Require().NoError(err)
	s.Nil(provider)
}
