package persistence_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestSamlIDPMetadataPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlIDPMetadataPersisterSuite))
}

type samlIDPMetadataPersisterSuite struct {
	test.Suite
}

func (s *samlIDPMetadataPersisterSuite) TestGet_ReturnsNilWhenNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	metadata, err := s.Storage.GetSamlIDPMetadataPersister().Get(uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001"), uuid.Must(uuid.NewV4()))
	s.Require().NoError(err)
	s.Nil(metadata)
}
