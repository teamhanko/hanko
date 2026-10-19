package persistence_test

import (
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

const testSourceIP = "127.0.0.1"

func TestAuditLogPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(auditLogPersisterSuite))
}

type auditLogPersisterSuite struct {
	test.Suite
}

func (s *auditLogPersisterSuite) TestAuditLogPersister_CreateStoresLongUserAgent() {
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
