package models

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditLogActorUserId(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())

	t.Run("uses the public id", func(t *testing.T) {
		internalID := uuid.Must(uuid.NewV4())
		publicID := uuid.Must(uuid.NewV4())

		auditLog, err := NewAuditLog(AuditLogLoginSuccess, RequestMeta{}, nil, &User{ID: internalID, PublicID: &publicID}, nil, tenantID)
		require.NoError(t, err)
		require.NotNil(t, auditLog.ActorUserId)
		assert.Equal(t, publicID, *auditLog.ActorUserId)
	})

	t.Run("falls back to the internal id without a public id", func(t *testing.T) {
		internalID := uuid.Must(uuid.NewV4())

		auditLog, err := NewAuditLog(AuditLogLoginSuccess, RequestMeta{}, nil, &User{ID: internalID}, nil, tenantID)
		require.NoError(t, err)
		require.NotNil(t, auditLog.ActorUserId)
		assert.Equal(t, internalID, *auditLog.ActorUserId)
	})

	t.Run("no actor without a user", func(t *testing.T) {
		auditLog, err := NewAuditLog(AuditLogLoginFailure, RequestMeta{}, nil, nil, nil, tenantID)
		require.NoError(t, err)
		assert.Nil(t, auditLog.ActorUserId)
	})
}
