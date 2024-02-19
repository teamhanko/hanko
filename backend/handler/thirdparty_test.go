package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	jwk2 "github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/suite"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"github.com/teamhanko/hanko/backend/utils"
)

func TestThirdPartySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(thirdPartySuite))
}

type thirdPartySuite struct {
	test.Suite
}

func (s *thirdPartySuite) setUpContext(request *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	rec := httptest.NewRecorder()
	c := e.NewContext(request, rec)
	return c, rec
}

func (s *thirdPartySuite) setUpHandler(cfg *config.Config) *ThirdPartyHandler {
	auditLogger := auditlog.NewLogger(s.Storage, cfg.AuditLog)

	jwkMngr, err := jwk.NewDefaultManager(cfg.Secrets.Keys, s.Storage.GetJwkPersister())
	s.Require().NoError(err)

	sessionMngr, err := session.NewManager(jwkMngr, *cfg)
	s.Require().NoError(err)

	handler := NewThirdPartyHandler(cfg, s.Storage, sessionMngr, auditLogger)
	return handler
}

func (s *thirdPartySuite) setUpConfig(enabledProviders []string, allowedRedirectURLs []string) *config.Config {
	cfg := &config.Config{
		ThirdParty: config.ThirdParty{
			Providers: config.ThirdPartyProviders{
				Apple: config.ThirdPartyProvider{
					Enabled:      false,
					ClientID:     "fakeClientID",
					Secret:       "fakeClientSecret",
					AllowLinking: true,
				},
				Google: config.ThirdPartyProvider{
					Enabled:      false,
					ClientID:     "fakeClientID",
					Secret:       "fakeClientSecret",
					AllowLinking: true,
				},
				GitHub: config.ThirdPartyProvider{
					Enabled:      false,
					ClientID:     "fakeClientID",
					Secret:       "fakeClientSecret",
					AllowLinking: true,
				},
				Discord: config.ThirdPartyProvider{
					Enabled:      false,
					ClientID:     "fakeClientID",
					Secret:       "fakeClientSecret",
					AllowLinking: true,
				},
			},
			ErrorRedirectURL:    "https://error.test.example",
			RedirectURL:         "https://api.test.example/callback",
			AllowedRedirectURLS: allowedRedirectURLs,
		},
		Secrets: config.Secrets{
			Keys: []string{"thirty-two-byte-long-test-secret"},
		},
		AuditLog: config.AuditLog{
			Storage: config.AuditLogStorage{Enabled: true},
		},
		Emails: config.Emails{
			MaxNumOfAddresses: 5,
		},
		Account: config.Account{
			AllowSignup: true,
		},
	}

	for _, provider := range enabledProviders {
		switch provider {
		case "apple":
			cfg.ThirdParty.Providers.Apple.Enabled = true
		case "google":
			cfg.ThirdParty.Providers.Google.Enabled = true
		case "github":
			cfg.ThirdParty.Providers.GitHub.Enabled = true
		case "discord":
			cfg.ThirdParty.Providers.Discord.Enabled = true
		}
	}

	err := cfg.PostProcess()
	s.Require().NoError(err)

	return cfg
}

func (s *thirdPartySuite) setUpFakeJwkSet() jwk2.Set {
	generator := test.JwkManager{}
	keySet, err := generator.GetPublicKeys()
	s.Require().NoError(err)
	return keySet
}

func (s *thirdPartySuite) setUpAppleIdToken(sub, aud, email string, emailVerified bool) string {
	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, sub)
	_ = token.Set(jwt.IssuedAtKey, time.Now().UTC())
	_ = token.Set(jwt.IssuerKey, "https://appleid.apple.com")
	_ = token.Set(jwt.AudienceKey, aud)
	_ = token.Set("email_verified", strconv.FormatBool(emailVerified))
	_ = token.Set("email", email)

	generator := test.JwkManager{}
	signingKey, err := generator.GetSigningKey()
	s.Require().NoError(err)

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, signingKey))
	s.Require().NoError(err)

	return string(signedToken)
}

func (s *thirdPartySuite) assertLocationHeaderHasToken(rec *httptest.ResponseRecorder) {
	location, err := url.Parse(rec.Header().Get("Location"))
	s.NoError(err)
	s.True(location.Query().Has(utils.HankoTokenQuery))
	s.NotEmpty(location.Query().Get(utils.HankoTokenQuery))
}

func (s *thirdPartySuite) assertStateCookieRemoved(rec *httptest.ResponseRecorder) {
	cookies := rec.Result().Cookies()
	s.Len(cookies, 1)
	s.Equal(utils.HankoThirdpartyStateCookie, cookies[0].Name)
	s.Equal(-1, cookies[0].MaxAge)
}
