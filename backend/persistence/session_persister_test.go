package persistence_test

import (
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

const (
	longUserAgentCharacter = "a"
	longUserAgentLength    = 278
)

func TestSessionPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(sessionPersisterSuite))
}

type sessionPersisterSuite struct {
	test.Suite
}

func (s *sessionPersisterSuite) TestSessionPersister_CreateStoresLongUserAgent() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	user := models.NewUser(tenantID)
	err := s.Storage.GetUserPersister().Create(user)
	s.Require().NoError(err)
	userAgent := strings.Repeat(longUserAgentCharacter, longUserAgentLength)
	now := time.Now().UTC()
	session := models.Session{
		ID:        uuid.Must(uuid.NewV4()),
		UserID:    user.ID,
		TenantID:  tenantID,
		UserAgent: nulls.NewString(userAgent),
		CreatedAt: now,
		UpdatedAt: now,
		LastUsed:  now,
	}
	err = s.Storage.GetSessionPersister().Create(session)
	s.Require().NoError(err)
	stored, err := s.Storage.GetSessionPersister().Get(session.ID, tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(stored)
	s.Equal(userAgent, stored.UserAgent.String)
}
