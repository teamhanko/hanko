package handler

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestOIDCSuite(t *testing.T) {
	suite.Run(t, new(oidcSuite))
}

type oidcSuite struct {
	test.Suite
}

func (s *oidcSuite) TestOIDCHandler_Paths() {
	cfg := &test.DefaultConfig
	cfg.OIDC.Enabled = true

	persister := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handlers := NewOIDCHandler(cfg, persister, nil, nil)

	s.Equal("/oauth/authorize", handlers.provider.AuthorizationEndpoint().Relative())
	s.Equal("/oauth/device_authorization", handlers.provider.DeviceAuthorizationEndpoint().Relative())
	s.Equal("/oauth/end_session", handlers.provider.EndSessionEndpoint().Relative())
	s.Equal("/oauth/introspect", handlers.provider.IntrospectionEndpoint().Relative())
	s.Equal("/oauth/keys", handlers.provider.KeysEndpoint().Relative())
	s.Equal("/oauth/revoke", handlers.provider.RevocationEndpoint().Relative())
	s.Equal("/oauth/token", handlers.provider.TokenEndpoint().Relative())
	s.Equal("/oauth/userinfo", handlers.provider.UserinfoEndpoint().Relative())
}

func (s *oidcSuite) TestOIDCHandler_well_known() {
	cfg := &test.DefaultConfig
	cfg.OIDC.Enabled = true

	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := NewPublicRouter(cfg, s.Storage, nil)
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	s.Require().NoError(err)

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	s.Equal("https://example.hanko.io", data["issuer"])
	s.Equal("https://example.hanko.io/oauth/authorize", data["authorization_endpoint"])
}

func (s *oidcSuite) TestOIDCHandler_authorize() {
	cfg := &test.DefaultConfig
	cfg.OIDC.Enabled = true

	err := s.LoadFixtures("../test/fixtures/oidc")
	s.Require().NoError(err)

	// Request generator: https://zitadel.com/docs/apis/openidoauth/authrequest
	path := "/oauth/authorize?client_id=19286ac4-2216-44dd-bb21-02a41ea3548d&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&scope=openid%20email%20profile%20offline_access&code_challenge=iMnq5o6zALKXGivsnlom_0F5_WYda32GHkxlV7mq7hQ&code_challenge_method=S256"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rev := httptest.NewRecorder()

	e := NewPublicRouter(cfg, s.Storage, nil)
	e.ServeHTTP(rev, req)

	s.Equal(http.StatusFound, rev.Code)

	// TODO: check redirect location
	s.True(strings.HasPrefix(rev.Header().Get("Location"), "/login/username?authRequestID="))

	uri, err := url.Parse(rev.Header().Get("Location"))
	s.Require().NoError(err)

	authRequestId := uri.Query().Get("authRequestID")

	// This is the tricky bit now - we need to simulate a login flow
	// Because Hanko has no built-in redirects in the login flow - we might need to do this client side or add a
	// redirect parameter somewhere with a custom redirect function.
	path = "/oauth/login?id=" + authRequestId
	req = httptest.NewRequest(http.MethodGet, path, nil)
	rev = httptest.NewRecorder()

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	s.Require().NoError(err)
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig)
	s.Require().NoError(err)
	token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"))
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)
	req.AddCookie(cookie)

	e.ServeHTTP(rev, req)

	s.Equal(http.StatusFound, rev.Code)

	// Let's follow the redirect from login
	path = rev.Header().Get("Location")
	req = httptest.NewRequest(http.MethodGet, path, nil)
	rev = httptest.NewRecorder()

	e.ServeHTTP(rev, req)

	s.Equal(http.StatusFound, rev.Code)
	s.True(strings.HasPrefix(rev.Header().Get("Location"), "http://localhost:8080/callback?code="))

	uri, err = url.Parse(rev.Header().Get("Location"))
	s.Require().NoError(err)

	code := uri.Query().Get("code")

	// This is now back with the client - let's simulate the token exchange
	path = "/oauth/token?grant_type=authorization_code&code=" + code + "&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback"
	req = httptest.NewRequest(http.MethodGet, path, nil)
	req.SetBasicAuth("19286ac4-2216-44dd-bb21-02a41ea3548d", "104cff48ae574505874884973de1f2488b8cd56ea55fdd45b2649a071af94617")
	rev = httptest.NewRecorder()

	e.ServeHTTP(rev, req)

	s.Equal(http.StatusOK, rev.Code)

	var data map[string]interface{}
	err = json.Unmarshal(rev.Body.Bytes(), &data)
	s.Require().NoError(err)

	s.NotEmpty(data["access_token"])
	s.NotEmpty(data["refresh_token"])

	// And let's also check out the userinfo endpoint
	path = "/oauth/introspect"
	req = httptest.NewRequest(http.MethodPost, path, strings.NewReader("token="+data["access_token"].(string)))
	req.SetBasicAuth("19286ac4-2216-44dd-bb21-02a41ea3548d", "104cff48ae574505874884973de1f2488b8cd56ea55fdd45b2649a071af94617")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rev = httptest.NewRecorder()

	e.ServeHTTP(rev, req)

	s.Equal(http.StatusOK, rev.Code)

	var introspect map[string]interface{}

	err = json.Unmarshal(rev.Body.Bytes(), &introspect)
	s.Require().NoError(err)

	s.Equal(introspect["active"], true)
	s.Equal(introspect["scope"], "openid email profile offline_access")
	s.Equal(introspect["client_id"], "19286ac4-2216-44dd-bb21-02a41ea3548d")
	s.Equal(introspect["sub"], "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")
	s.Equal(introspect["email"], "john.doe@example.com")
	s.Equal(introspect["email_verified"], true)
}
