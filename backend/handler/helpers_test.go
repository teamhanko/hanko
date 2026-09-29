package handler

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
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

// generateSessionWithToken builds a session JWT and its backing session row, taking the
// internal id separately from the public id since the session row's UserID is a real
// foreign key to users.id, while the JWT subject is the tenant-scoped public_id - the
// only identifier a real client would ever see.
func generateSessionWithToken(storage persistence.Persister, internalUserID, publicUserID uuid.UUID, tenantID uuid.UUID, expiresAt *time.Time) (string, uuid.UUID, error) {
	manager := getDefaultSessionManager(storage)
	userJWT := dto.UserJWT{
		UserID: publicUserID.String(),
	}

	var token string
	var rawToken jwt.Token
	var err error
	if expiresAt != nil {
		token, rawToken, err = manager.GenerateJWT(userJWT, tenantID, session.WithValue(jwt.ExpirationKey, expiresAt))
	} else {
		token, rawToken, err = manager.GenerateJWT(userJWT, tenantID)
	}
	if err != nil {
		return "", uuid.Nil, err
	}

	sessionIDClaim, _ := rawToken.Get("session_id")
	sessionID := uuid.FromStringOrNil(sessionIDClaim.(string))
	err = storage.GetSessionPersister().Create(models.Session{
		ID:        sessionID,
		UserID:    internalUserID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
		LastUsed:  time.Now().UTC(),
		TenantID:  tenantID,
	})
	if err != nil {
		return "", uuid.Nil, err
	}

	return token, sessionID, nil
}

func generateSessionWithCookie(storage persistence.Persister, internalUserID, publicUserID uuid.UUID, tenantID uuid.UUID, expiresAt *time.Time) (*http.Cookie, error) {
	token, _, err := generateSessionWithToken(storage, internalUserID, publicUserID, tenantID, expiresAt)
	if err != nil {
		return nil, err
	}
	return getDefaultSessionManager(storage).GenerateCookie(token)
}
