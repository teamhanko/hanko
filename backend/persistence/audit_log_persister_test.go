package persistence_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/pop/v6"
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

const backfillActorUserIdMigrationVersion = "20261016120000"

// runBackfillActorUserIdMigration runs only the up or down direction of the
// change_audit_logs_actor_user_id_to_public_id migration, so the test doesn't depend on it being
// the newest migration (unlike MigrateDown(1), which always reverts the newest one).
func (s *auditLogPersisterSuite) runBackfillActorUserIdMigration(direction string) {
	box, err := pop.NewMigrationBox(os.DirFS("migrations"), s.Storage.GetConnection())
	s.Require().NoError(err)

	migrations := box.UpMigrations.Migrations
	if direction == "down" {
		migrations = box.DownMigrations.Migrations
	}
	for _, migration := range migrations {
		if migration.Version == backfillActorUserIdMigrationVersion {
			s.Require().NoError(migration.Run(s.Storage.GetConnection()))
			return
		}
	}
	s.FailNow("backfill migration not found", "direction %s", direction)
}

// TestAuditLogPersister_BackfillActorUserIdToPublicID confirms the
// change_audit_logs_actor_user_id_to_public_id migration rewrites actor_user_id from the internal
// user id to the public id - only for users whose public id differs - and that down reverts it.
func (s *auditLogPersisterSuite) TestAuditLogPersister_BackfillActorUserIdToPublicID() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	now := time.Now().UTC()

	divergedInternalID := uuid.Must(uuid.NewV4())
	divergedPublicID := uuid.Must(uuid.NewV4())
	legacyID := uuid.Must(uuid.NewV4())
	deletedUserID := uuid.Must(uuid.NewV4())

	userPersister := s.Storage.GetUserPersister()
	s.Require().NoError(userPersister.Create(models.User{ID: divergedInternalID, PublicID: &divergedPublicID, TenantID: tenantID, CreatedAt: now, UpdatedAt: now}))
	s.Require().NoError(userPersister.Create(models.User{ID: legacyID, PublicID: &legacyID, TenantID: tenantID, CreatedAt: now, UpdatedAt: now}))

	// Seed rows the way they were written before the fix: with the internal user id.
	newAuditLog := func(actorUserID *uuid.UUID) models.AuditLog {
		auditLog := models.AuditLog{
			ID:                uuid.Must(uuid.NewV4()),
			Type:              models.AuditLogLoginSuccess,
			MetaHttpRequestId: uuid.Must(uuid.NewV4()).String(),
			MetaSourceIp:      testSourceIP,
			ActorUserId:       actorUserID,
			TenantID:          tenantID,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		s.Require().NoError(s.Storage.GetAuditLogPersister().Create(auditLog))
		return auditLog
	}

	divergedLog := newAuditLog(&divergedInternalID)
	legacyLog := newAuditLog(&legacyID)
	deletedUserLog := newAuditLog(&deletedUserID)
	anonymousLog := newAuditLog(nil)

	actorUserID := func(auditLog models.AuditLog) *uuid.UUID {
		stored, err := s.Storage.GetAuditLogPersister().Get(auditLog.ID, tenantID)
		s.Require().NoError(err)
		s.Require().NotNil(stored)
		return stored.ActorUserId
	}

	s.runBackfillActorUserIdMigration("up")

	s.Equal(divergedPublicID, *actorUserID(divergedLog))
	s.Equal(legacyID, *actorUserID(legacyLog))
	s.Equal(deletedUserID, *actorUserID(deletedUserLog), "logs of deleted users can't be backfilled and stay unchanged")
	s.Nil(actorUserID(anonymousLog))

	s.runBackfillActorUserIdMigration("down")

	s.Equal(divergedInternalID, *actorUserID(divergedLog))
	s.Equal(legacyID, *actorUserID(legacyLog))
	s.Equal(deletedUserID, *actorUserID(deletedUserLog))
	s.Nil(actorUserID(anonymousLog))
}
