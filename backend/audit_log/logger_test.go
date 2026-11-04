package auditlog_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	auditlog "github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestLoggerSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(loggerSuite))
}

type loggerSuite struct {
	test.Suite
}

func (s *loggerSuite) createAndGetActorUserID(user *models.User) *uuid.UUID {
	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	logger := auditlog.NewLogger(s.Storage, config.AuditLog{Storage: config.AuditLogStorage{Enabled: true}})

	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	s.Require().NoError(logger.Create(c, models.AuditLogLoginSuccess, user, nil, tenantID))

	auditLogs, err := s.Storage.GetAuditLogPersister().List(1, 100, nil, nil, nil, "", "", "", "", tenantID)
	s.Require().NoError(err)
	s.Require().Len(auditLogs, 1)
	return auditLogs[0].ActorUserId
}

func (s *loggerSuite) TestCreate_ResolvesPublicIDOfUserWithInternalIDOnly() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	internalID := uuid.Must(uuid.NewV4())
	publicID := uuid.Must(uuid.NewV4())
	now := time.Now().UTC()
	s.Require().NoError(s.Storage.GetUserPersister().Create(models.User{
		ID:        internalID,
		PublicID:  &publicID,
		TenantID:  uuid.FromStringOrNil(config.DefaultTenantID),
		CreatedAt: now,
		UpdatedAt: now,
	}))

	actorUserID := s.createAndGetActorUserID(&models.User{ID: internalID})
	s.Require().NotNil(actorUserID)
	s.Equal(publicID, *actorUserID)
}

func (s *loggerSuite) TestCreate_KeepsInternalIDOfUnknownUser() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	internalID := uuid.Must(uuid.NewV4())

	actorUserID := s.createAndGetActorUserID(&models.User{ID: internalID})
	s.Require().NotNil(actorUserID)
	s.Equal(internalID, *actorUserID)
}
