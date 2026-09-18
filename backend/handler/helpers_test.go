package handler

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/session"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func getDefaultSessionManager(storage persistence.Persister) session.Manager {
	jwkManager, _ := local_db.NewDefaultManager(test.DefaultConfig.SecretKeys, storage.GetJwkPersister())
	sessionManager, _ := session.NewManager(jwkManager, test.DefaultConfig.TenantConfig)
	return sessionManager
}

// generateSigningKeyForTenant mirrors what handler/tenant.go's Create
// does for a real tenant - needed because fixture-inserted tenants skip it.
func generateSigningKeyForTenant(storage persistence.Persister, tenantID uuid.UUID) error {
	manager, err := local_db.NewDefaultManager(test.DefaultConfig.SecretKeys, storage.GetJwkPersister())
	if err != nil {
		return err
	}
	_, err = manager.GenerateKey(tenantID)
	return err
}

func generateSessionCookie(storage persistence.Persister, userId uuid.UUID, tenantID uuid.UUID) (*http.Cookie, error) {
	manager := getDefaultSessionManager(storage)
	token, rawToken, err := manager.GenerateJWT(dto.UserJWT{
		UserID: userId.String(),
	}, tenantID)
	if err != nil {
		return nil, err
	}
	sessionID, _ := rawToken.Get("session_id")
	_ = storage.GetSessionPersister().Create(models.Session{
		ID:        uuid.FromStringOrNil(sessionID.(string)),
		UserID:    userId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: nil,
		LastUsed:  time.Now(),
		TenantID:  tenantID,
	})
	cookie, err := manager.GenerateCookie(token)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}
