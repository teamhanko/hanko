package persistence_test

import (
	"testing"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestSamlStatePersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlStatePersisterSuite))
}

type samlStatePersisterSuite struct {
	test.Suite
}

// GetByNonce makes a second reader of the same state wait until the first reader's transaction ends.
func (s *samlStatePersisterSuite) TestGetByNonce_LocksUntilTransactionEnds() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)

	state, err := models.NewSamlState("test-nonce", "encrypted-state-blob", tenantID)
	s.Require().NoError(err)
	s.Require().NoError(s.Storage.GetSamlStatePersister().Create(*state))

	secondRead := make(chan error, 1)
	secondReadReturnedEarly := false

	err = s.Storage.Transaction(func(tx *pop.Connection) error {
		statePersister := s.Storage.GetSamlStatePersisterWithConnection(tx)

		lockedState, err := statePersister.GetByNonce("test-nonce", tenantID)
		if err != nil {
			return err
		}

		go func() {
			secondRead <- s.Storage.Transaction(func(tx *pop.Connection) error {
				_, err := s.Storage.GetSamlStatePersisterWithConnection(tx).GetByNonce("test-nonce", tenantID)
				return err
			})
		}()

		select {
		case <-secondRead:
			secondReadReturnedEarly = true
		case <-time.After(500 * time.Millisecond):
		}

		return statePersister.Delete(*lockedState)
	})
	s.Require().NoError(err)
	s.Require().False(secondReadReturnedEarly, "the second read did not wait for the first transaction to end")

	// The first transaction deleted the state, so the second read finds nothing once it is released.
	s.Require().Error(<-secondRead)
}

// GetByNonce refuses to run outside a transaction.
func (s *samlStatePersisterSuite) TestGetByNonce_RequiresTransaction() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	_, err := s.Storage.GetSamlStatePersister().GetByNonce("test-nonce", uuid.FromStringOrNil(config.DefaultTenantID))
	s.Require().ErrorContains(err, "must be called inside a transaction")
}
