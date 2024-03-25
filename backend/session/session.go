package session

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	hankoJwk "github.com/teamhanko/hanko/backend/crypto/jwk"
	hankoJwt "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"net/http"
	"time"
)

type Manager interface {
	GenerateJWT(userId uuid.UUID, userDto *dto.EmailJwt) (string, error)
	Verify(string) (jwt.Token, error)
	GenerateCookie(token string) (*http.Cookie, error)
	DeleteCookie() (*http.Cookie, error)
}

// Manager is used to create and verify session JWTs
type manager struct {
	jwtGenerator  hankoJwt.Generator
	sessionLength time.Duration
	cookieConfig  cookieConfig
	issuer        string
	audience      []string
}

type cookieConfig struct {
	Name     string
	Domain   string
	HttpOnly bool
	SameSite http.SameSite
	Secure   bool
}

const (
	GeneratorCreateFailure = "failed to create session generator: %w"
)

// NewManager returns a new Manager which will be used to create and verify sessions JWTs
func NewManager(jwkManager hankoJwk.Manager, config config.Config) (Manager, error) {
	signatureKey, err := jwkManager.GetSigningKey()
	if err != nil {
		return nil, fmt.Errorf(GeneratorCreateFailure, err)
	}
	verificationKeys, err := jwkManager.GetPublicKeys()
	if err != nil {
		return nil, fmt.Errorf(GeneratorCreateFailure, err)
	}
	g, err := hankoJwt.NewGenerator(signatureKey, verificationKeys)
	if err != nil {
		return nil, fmt.Errorf(GeneratorCreateFailure, err)
	}

	duration, _ := time.ParseDuration(config.Session.Lifespan) // error can be ignored, value is checked in config validation
	sameSite := http.SameSite(0)
	switch config.Session.Cookie.SameSite {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}
	var audience []string
	if config.Session.Audience != nil && len(config.Session.Audience) > 0 {
		audience = config.Session.Audience
	} else {
		audience = []string{config.Webauthn.RelyingParty.Id}
	}

	return &manager{
		jwtGenerator:  g,
		sessionLength: duration,
		issuer:        config.Session.Issuer,
		cookieConfig: cookieConfig{
			Name:     config.Session.Cookie.GetName(),
			Domain:   config.Session.Cookie.Domain,
			HttpOnly: config.Session.Cookie.HttpOnly,
			SameSite: sameSite,
			Secure:   config.Session.Cookie.Secure,
		},
		audience: audience,
	}, nil
}

// GenerateJWT creates a new session JWT for the given user
func (m *manager) GenerateJWT(userId uuid.UUID, email *dto.EmailJwt) (string, error) {
	issuedAt := time.Now()
	expiration := issuedAt.Add(m.sessionLength)

	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, userId.String())
	_ = token.Set(jwt.IssuedAtKey, issuedAt)
	_ = token.Set(jwt.ExpirationKey, expiration)
	_ = token.Set(jwt.AudienceKey, m.audience)

	if email != nil {
		_ = token.Set("email", &email)
	}

	if m.issuer != "" {
		_ = token.Set(jwt.IssuerKey, m.issuer)
	}

	signed, err := m.jwtGenerator.Sign(token)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

// Verify verifies the given JWT and returns a parsed one if verification was successful
func (m *manager) Verify(token string) (jwt.Token, error) {
	parsedToken, err := m.jwtGenerator.Verify([]byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to verify session token: %w", err)
	}

	return parsedToken, nil
}

// GenerateCookie creates a new session cookie for the given user
func (m *manager) GenerateCookie(token string) (*http.Cookie, error) {
	return &http.Cookie{
		Name:     m.cookieConfig.Name,
		Value:    token,
		Domain:   m.cookieConfig.Domain,
		Path:     "/",
		Secure:   m.cookieConfig.Secure,
		HttpOnly: m.cookieConfig.HttpOnly,
		SameSite: m.cookieConfig.SameSite,
		MaxAge:   int(m.sessionLength.Seconds()),
	}, nil
}

// DeleteCookie returns a cookie that will expire the cookie on the frontend
func (m *manager) DeleteCookie() (*http.Cookie, error) {
	return &http.Cookie{
		Name:     m.cookieConfig.Name,
		Value:    "",
		Domain:   m.cookieConfig.Domain,
		Path:     "/",
		Secure:   m.cookieConfig.Secure,
		HttpOnly: m.cookieConfig.HttpOnly,
		SameSite: m.cookieConfig.SameSite,
		MaxAge:   -1,
	}, nil
}
