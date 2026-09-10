package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestPublicIdConsistencySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(publicIdConsistencySuite))
}

type publicIdConsistencySuite struct {
	test.Suite
}

// TestPublicIdConsistency verifies the cross-surface consistency the plan requires: the session
// JWT's sub claim, /me, and POST /sessions/validate must all agree on a single identifier for a
// given user - and that identifier must be public_id, never the internal id. (Admin API coverage
// for this lives separately in TestUserHandlerAdmin_Get_UsesPublicIdNotInternalId.)
func (s *publicIdConsistencySuite) TestPublicIdConsistency() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/public_id_consistency")
	s.Require().NoError(err)

	internalID := uuid.FromStringOrNil("7c3c8f0a-5555-4a11-8a11-555555555555")
	publicID := uuid.FromStringOrNil("8d4d9f1b-6666-4b22-8b22-666666666666")
	tenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	cfg := test.DefaultConfig
	s.Require().NoError(cfg.PostProcess())
	s.Require().NoError(local_db.SyncSecretKeys(&cfg, s.Storage))

	// 1. Generate a real session the same way the login flow does: fetch the user model and feed
	// it through UserJWTFromUserModel, the exact function that now sources sub from PublicID.
	user, err := s.Storage.GetUserPersister().Get(internalID, tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(user)

	manager := getDefaultSessionManager(s.Storage)
	token, rawToken, err := manager.GenerateJWT(dto.UserJWTFromUserModel(user), tenantID)
	s.Require().NoError(err)

	// 2. JWT sub claim must be public_id, never the internal id.
	s.Equal(publicID.String(), rawToken.Subject(), "JWT sub claim must be public_id")

	sessionID, _ := rawToken.Get("session_id")
	sessionUUID := uuid.FromStringOrNil(sessionID.(string))
	now := time.Now().UTC()
	s.Require().NoError(s.Storage.GetSessionPersister().Create(models.Session{
		ID:        sessionUUID,
		TenantID:  tenantID,
		UserID:    internalID,
		CreatedAt: now,
		UpdatedAt: now,
		LastUsed:  now,
	}))

	cookie, err := manager.GenerateCookie(token)
	s.Require().NoError(err)

	publicRouter := NewPublicRouter(&cfg, s.Storage, nil, nil)

	// 3. /me must show public_id as both "id" and "user_id".
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	publicRouter.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code)

	var profile dto.ProfileData
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &profile))
	s.Equal(publicID, profile.ID, "/me must show public_id as \"id\"")
	s.Equal(publicID, profile.UserID, "/me must show public_id as \"user_id\"")

	// 4. POST /sessions/validate must return public_id as user_id too - this is the endpoint an
	// external system uses to ask "who does this token belong to", so it must agree with /me.
	req = httptest.NewRequest(http.MethodGet, "/sessions/validate", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	publicRouter.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code)

	var validateResponse dto.ValidateSessionResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &validateResponse))
	s.True(validateResponse.IsValid)
	s.Require().NotNil(validateResponse.UserID)
	s.Equal(publicID, *validateResponse.UserID, "/sessions/validate must return public_id as user_id")
}
