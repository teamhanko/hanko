package handler

import (
	"net/http"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/session"
	"github.com/teamhanko/hanko/backend/v2/test"
)

func getDefaultSessionManager(storage persistence.Persister) session.Manager {
	jwkManager, _ := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, storage.GetJwkPersister())
	sessionManager, _ := session.NewManager(jwkManager, test.DefaultConfig)
	return sessionManager
}

func generateSessionCookie(storage persistence.Persister, userId uuid.UUID) (*http.Cookie, error) {
	manager := getDefaultSessionManager(storage)
	token, rawToken, err := manager.GenerateJWT(dto.UserJWT{
		UserID: userId.String(),
	})
	if err != nil {
		return nil, err
	}
	sessionID, _ := rawToken.Get("session_id")
	_ = storage.GetSessionPersister().Create(models.Session{
		ID:        uuid.FromStringOrNil(sessionID.(string)),
		UserID:    nulls.NewUUID(userId),
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
