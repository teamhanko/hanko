package handler

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/handler/oidc"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/zitadel/oidc/v2/pkg/op"
	"golang.org/x/text/language"
	"net/http"
)

type OIDCHandler struct {
	cfg            *config.Config
	persister      persistence.Persister
	sessionManager session.Manager
	auditLogger    auditlog.Logger
	provider       op.OpenIDProvider
}

func NewOIDCHandler(
	cfg *config.Config,
	persister persistence.Persister,
	sessionManager session.Manager,
	auditLogger auditlog.Logger,
) *OIDCHandler {
	if !cfg.OIDC.Enabled {
		return nil
	}

	key, err := base64.URLEncoding.DecodeString(cfg.OIDC.Key)
	if err != nil {
		panic(err)
	}

	if len(key) != 32 {
		panic("key must be 32 bytes long")
	}

	pathLoggedOut := "/logged_out"

	var extraOptions []op.Option

	config := &op.Config{
		CryptoKey: [32]byte(key),

		// will be used if the end_session endpoint is called without a post_logout_redirect_uri
		DefaultLogoutRedirectURI: pathLoggedOut,

		// enables code_challenge_method S256 for PKCE (and therefore PKCE in general)
		CodeMethodS256: true,

		// enables additional client_id/client_secret authentication by form post (not only HTTP Basic Auth)
		AuthMethodPost: true,

		// enables additional authentication by using private_key_jwt
		AuthMethodPrivateKeyJWT: false,

		// enables refresh_token grant use
		GrantTypeRefreshToken: true,

		// enables use of the `request` Object parameter
		RequestObjectSupported: true,

		// this example has only static texts (in English), so we'll set the here accordingly
		SupportedUILocales: []language.Tag{language.English},
	}

	storage := oidc.NewStorage(persister)
	for _, client := range cfg.OIDC.Clients {
		err := storage.AddClient(&client)
		if err != nil {
			panic(err)
		}
	}

	provider, err := op.NewOpenIDProvider(cfg.OIDC.Issuer, config, storage, append([]op.Option{
		op.WithCustomEndpoints(
			op.NewEndpoint("/oauth/authorize"),
			op.NewEndpoint("/oauth/token"),
			op.NewEndpoint("/oauth/userinfo"),
			op.NewEndpoint("/oauth/revoke"),
			op.NewEndpoint("/oauth/end_session"),
			op.NewEndpoint("/oauth/keys"),
		),
		op.WithCustomDeviceAuthorizationEndpoint(op.NewEndpoint("/oauth/device_authorization")),
	}, extraOptions...)...)
	if err != nil {
		panic(err)
	}

	fmt.Println("OIDC provider initialized")
	f := op.AuthCallbackURL(provider)
	fmt.Println("OIDC callback url:", f(context.Background(), "testID"))
	fmt.Println("OIDC callback url:")

	return &OIDCHandler{
		cfg:            cfg,
		persister:      persister,
		sessionManager: sessionManager,
		auditLogger:    auditLogger,
		provider:       provider,
	}
}

func (h *OIDCHandler) Handler(c echo.Context) error {
	h.provider.HttpHandler().ServeHTTP(c.Response(), c.Request())

	return nil
}

func (h *OIDCHandler) LoginHandler(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	authRequestID := c.QueryParam("id")
	if authRequestID == "" {
		return c.String(400, "id parameter missing")
	}

	uid, err := uuid.FromString(authRequestID)
	if err != nil {
		return c.String(400, "id parameter invalid")
	}

	persister := h.persister.GetOIDCAuthRequestPersister()

	err = persister.AuthorizeUser(c.Request().Context(), uid, sessionToken.Subject())
	if err != nil {
		return c.String(500, "error authorizing user")
	}

	return c.Redirect(http.StatusFound, "/oauth/authorize/callback?id="+authRequestID)
}
