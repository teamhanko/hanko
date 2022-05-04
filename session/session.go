package session

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/config"
	hankoJwk "github.com/teamhanko/hanko/crypto/jwk"
	hankoJwt "github.com/teamhanko/hanko/crypto/jwt"
	"net/http"
	"time"
)

type Manager interface {
	GenerateJWT(uuid.UUID) (string, error)
	Verify(string) (jwt.Token, error)
	GenerateCookie(userId uuid.UUID) (*http.Cookie, error)
}

// Manager is used to create and verify session JWTs
type manager struct {
	jwtGenerator  hankoJwt.Generator
	sessionLength time.Duration
	cookieConfig   config.Cookie
}

// NewManager returns a new Manager which will be used to create and verify sessions JWTs
func NewManager(jwkManager hankoJwk.Manager, config config.Cookie) (Manager, error) {
	signatureKey, err := jwkManager.GetSigningKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	verificationKeys, err := jwkManager.GetPublicKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	g, err := hankoJwt.NewGenerator(signatureKey, verificationKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	return &manager{
		jwtGenerator:  g,
		sessionLength: time.Minute * 60, // TODO: should come from config
	}, nil
}

// Generate creates a new session JWT for the given user
func (g *manager) GenerateJWT(userId uuid.UUID) (string, error) {
	issuedAt := time.Now()
	expiration := issuedAt.Add(g.sessionLength)

	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, userId.String())
	_ = token.Set(jwt.IssuedAtKey, issuedAt)
	_ = token.Set(jwt.ExpirationKey, expiration)
	//_ = token.Set(jwt.AudienceKey, []string{"http://localhost"})

	signed, err := g.jwtGenerator.Sign(token)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

// Verify verifies the given JWT and returns a parsed one if verification was successful
func (g *manager) Verify(token string) (jwt.Token, error) {
	parsedToken, err := g.jwtGenerator.Verify([]byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to verify session token: %w", err)
	}

	return parsedToken, nil
}

func (g *manager) GenerateCookie(userId uuid.UUID) (*http.Cookie, error) {
	jwt, err := g.GenerateJWT(userId)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     "hanko",
		Value:    jwt,
		Domain:   g.cookieConfig.Domain,
		Secure:   true,
		HttpOnly: g.cookieConfig.HttpOnly,
		//TODO: config has the SameSite Parameter which is string, http.SameSite is int do we need to make this configurable?
		SameSite: http.SameSiteLaxMode,
	}, nil
}
