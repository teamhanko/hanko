package handler

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"time"
)

func getDefaultSessionManager(storage persistence.Persister) session.Manager {
	jwkManager, _ := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, storage.GetJwkPersister())
	sessionManager, _ := session.NewManager(jwkManager, test.DefaultConfig)
	return sessionManager
}

func generateSessionCookie(storage persistence.Persister, userId uuid.UUID) (*http.Cookie, error) {
	manager := getDefaultSessionManager(storage)
	token, rawToken, err := manager.GenerateJWT(userId, nil)
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
	})
	cookie, err := manager.GenerateCookie(token)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}
