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
	testSourceIP           = "127.0.0.1"
)

func TestUserAgentSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(userAgentSuite))
}

type userAgentSuite struct {
	test.Suite
}

func (s *userAgentSuite) TestSessionPersister_CreateStoresLongUserAgent() {
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

func (s *userAgentSuite) TestAuditLogPersister_CreateStoresLongUserAgent() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	userAgent := strings.Repeat(longUserAgentCharacter, longUserAgentLength)
	now := time.Now().UTC()
	auditLog := models.AuditLog{
		ID:                uuid.Must(uuid.NewV4()),
		Type:              models.AuditLogLoginSuccess,
		MetaHttpRequestId: uuid.Must(uuid.NewV4()).String(),
		MetaSourceIp:      testSourceIP,
		MetaUserAgent:     userAgent,
		TenantID:          uuid.FromStringOrNil(config.DefaultTenantID),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.Storage.GetAuditLogPersister().Create(auditLog)
	s.Require().NoError(err)

	stored, err := s.Storage.GetAuditLogPersister().Get(auditLog.ID, auditLog.TenantID)
	s.Require().NoError(err)
	s.Require().NotNil(stored)
	s.Equal(userAgent, stored.MetaUserAgent)
}
