package main

import (
	"context"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"github.com/zitadel/oidc/v2/pkg/op"
	"gopkg.in/square/go-jose.v2"
	"sync"
	"time"
)

type Storage struct {
	lock sync.RWMutex

	db                     *pop.Connection
	clients                map[string]*Client
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration

	accessTokens  persistence.OIDCAccessTokensPersister
	refreshTokens persistence.OIDCRefreshTokensPersister
	authRequests  persistence.OIDCAuthRequestPersister
	keys          persistence.OIDCKeysPersister
	users         persistence.UserPersister
}

func NewStorage(db *pop.Connection) *Storage {
	return &Storage{
		db:            db,
		accessTokens:  persistence.NewOIDCAccessTokensPersister(db),
		refreshTokens: persistence.NewOIDCRefreshTokensPersister(db),
		authRequests:  persistence.NewOIDCAuthRequestPersister(db),
	}
}

// CreateAuthRequest implements the op.Storage interface
// it will be called after parsing and validation of the authentication request
func (s *Storage) CreateAuthRequest(ctx context.Context, req *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
	if len(req.Prompt) == 1 && req.Prompt[0] == "none" {
		// With prompt=none, there is no way for the user to log in
		// so return error right away.
		return nil, oidc.ErrLoginRequired()
	}

	// typically, you'll fill your storage / storage model with the information of the passed object
	request := authRequestToInternal(req, userID)

	// you'll also have to create a unique id for the request (this might be done by your database; we'll use a uuid)
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", err)
	}

	request.ID = uid

	// and save it in your database (for demonstration purposed we will use a simple map)
	err = s.authRequests.Create(ctx, request.ToModel())
	if err != nil {
		return nil, err
	}

	return request, nil
}

// AuthRequestByID implements the op.Storage interface
// it will be called after the Login UI redirects back to the OIDC endpoint
func (s *Storage) AuthRequestByID(ctx context.Context, id string) (op.AuthRequest, error) {
	uid, err := uuid.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("failed parse uuid: %w", err)
	}

	request, err := s.authRequests.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("could not get auth request: %w", err)
	}

	return NewAuthRequestFromModel(request)
}

// AuthRequestByCode implements the op.Storage interface
// it will be called after parsing and validation of the token request (in an authorization code flow)
func (s *Storage) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	request, err := s.authRequests.GetAuthRequestByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("could not get auth request by code: %w", err)
	}

	return NewAuthRequestFromModel(request)
}

// SaveAuthCode implements the op.Storage interface
// it will be called after the authentication has been successful and before redirecting the user agent to the
// redirect_uri (in an authorization code flow)
func (s *Storage) SaveAuthCode(ctx context.Context, id string, code string) error {
	uid, err := uuid.FromString(id)
	if err != nil {
		return fmt.Errorf("failed parse uuid: %w", err)
	}

	err = s.authRequests.StoreAuthCode(ctx, uid, code)
	if err != nil {
		return fmt.Errorf("could not store auth code: %w", err)
	}

	return nil
}

// DeleteAuthRequest implements the op.Storage interface
// it will be called after creating the token response (id and access tokens) for a valid
// - authentication request (in an implicit flow)
// - token request (in an authorization code flow)
func (s *Storage) DeleteAuthRequest(ctx context.Context, id string) error {
	uid, err := uuid.FromString(id)
	if err != nil {
		return fmt.Errorf("failed parse uuid: %w", err)
	}

	err = s.authRequests.Delete(ctx, uid)
	if err != nil {
		return fmt.Errorf("could not delete auth request: %w", err)
	}

	return nil
}

// createAccessToken will store an access_token in-memory based on the provided information
func (s *Storage) createAccessToken(ctx context.Context, clientID, subject string, refreshTokenID uuid.UUID, audience, scopes []string) (*models.AccessToken, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", err)
	}

	var refreshToken *models.RefreshToken
	if refreshTokenID != uuid.Nil {
		refreshToken = &models.RefreshToken{ID: refreshTokenID}
	}

	token := models.AccessToken{
		ID:           uid,
		ClientID:     clientID,
		RefreshToken: refreshToken,
		Subject:      subject,
		Audience:     audience,
		Expiration:   time.Now().Add(s.accessTokenExpiration),
		Scopes:       scopes,
	}

	err = s.accessTokens.Create(ctx, token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// createRefreshToken will store a refresh_token in-memory based on the provided information
func (s *Storage) createRefreshToken(ctx context.Context, accessToken *models.AccessToken, amr []string, authTime time.Time) (*models.RefreshToken, error) {
	token := models.RefreshToken{
		ID:         accessToken.RefreshToken.ID,
		AuthTime:   authTime,
		AMR:        amr,
		ClientID:   accessToken.ClientID,
		UserID:     accessToken.Subject,
		Audience:   accessToken.Audience,
		Expiration: time.Now().Add(s.refreshTokenExpiration),
		Scopes:     accessToken.Scopes,
	}

	err := s.refreshTokens.Create(ctx, token)
	if err != nil {
		return nil, err
	}

	return &token, err
}

// renewRefreshToken checks the provided refresh_token and creates a new one based on the current
func (s *Storage) renewRefreshToken(ctx context.Context, clientID, currentRefreshToken string) (*models.RefreshToken, error) {
	uid, err := uuid.FromString(currentRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uuid: %w", err)
	}

	token, err := s.refreshTokens.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	if token.ClientID != clientID {
		return nil, op.ErrInvalidRefreshToken
	}

	// deletes the refresh token and all access tokens which were issued based on this refresh token
	err = s.refreshTokens.Delete(ctx, *token)
	if err != nil {
		return nil, fmt.Errorf("failed to delete refresh token: %w", err)
	}

	// creates a new refresh token based on the current one
	uid, err = uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", err)
	}

	token.ID = uid

	err = s.refreshTokens.Create(ctx, *token)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return token, nil
}

func (s *Storage) exchangeRefreshToken(ctx context.Context, request op.TokenExchangeRequest) (accessTokenID string, newRefreshToken string, expiration time.Time, err error) {
	applicationID := request.GetClientID()
	authTime := request.GetAuthTime()

	refreshTokenID, err := uuid.NewV4()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate uuid: %w", err)
	}

	accessToken, err := s.createAccessToken(ctx, applicationID, request.GetSubject(), refreshTokenID, request.GetAudience(), request.GetScopes())
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshToken, err := s.createRefreshToken(ctx, accessToken, nil, authTime)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken.ID.String(), refreshToken.ID.String(), accessToken.Expiration, nil
}

// CreateAccessToken implements the op.Storage interface
// it will be called for all requests able to return an access token (Authorization Code Flow, Implicit Flow, JWT Profile, ...)
func (s *Storage) CreateAccessToken(ctx context.Context, request op.TokenRequest) (accessTokenID string, expiration time.Time, err error) {
	var applicationID string
	switch req := request.(type) {
	case *AuthRequest:
		// if authenticated for an app (auth code / implicit flow) we must save the client_id to the token
		applicationID = req.ApplicationID
	case op.TokenExchangeRequest:
		applicationID = req.GetClientID()
	default:
		panic("invalid state encountered")
	}

	token, err := s.createAccessToken(ctx, applicationID, request.GetSubject(), uuid.Nil, request.GetAudience(), request.GetScopes())
	if err != nil {
		return "", time.Time{}, err
	}

	return token.ID.String(), token.Expiration, nil
}

// CreateAccessAndRefreshTokens implements the op.Storage interface
// it will be called for all requests able to return an access and refresh token (Authorization Code Flow, Refresh Token Request)
func (s *Storage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshTokenID string, expiration time.Time, err error) {
	// generate tokens via token exchange flow if request is relevant
	if teReq, ok := request.(op.TokenExchangeRequest); ok {
		return s.exchangeRefreshToken(ctx, teReq)
	}

	// get the information depending on the request type / implementation
	applicationID, authTime, amr := getInfoFromRequest(request)

	// if currentRefreshToken is empty (Code Flow) we will have to create a new refresh token
	if currentRefreshToken == "" {
		refreshTokenID, err := uuid.NewV4()
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("failed to generate uuid: %w", err)
		}

		accessToken, err := s.createAccessToken(ctx, applicationID, request.GetSubject(), refreshTokenID, request.GetAudience(), request.GetScopes())
		if err != nil {
			return "", "", time.Time{}, err
		}

		refreshToken, err := s.createRefreshToken(ctx, accessToken, amr, authTime)
		if err != nil {
			return "", "", time.Time{}, err
		}

		return accessToken.ID.String(), refreshToken.ID.String(), accessToken.Expiration, nil
	}

	// if we get here, the currentRefreshToken was not empty, so the call is a refresh token request
	// we therefore will have to check the currentRefreshToken and renew the refresh token
	refreshToken, err := s.renewRefreshToken(ctx, applicationID, currentRefreshToken)
	if err != nil {
		return "", "", time.Time{}, err
	}

	accessToken, err := s.createAccessToken(ctx, applicationID, request.GetSubject(), refreshToken.ID, request.GetAudience(), request.GetScopes())
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken.ID.String(), refreshToken.ID.String(), accessToken.Expiration, nil
}

// TokenRequestByRefreshToken implements the op.Storage interface
// it will be called after parsing and validation of the refresh token request
func (s *Storage) TokenRequestByRefreshToken(ctx context.Context, refreshTokenID string) (op.RefreshTokenRequest, error) {
	uid, err := uuid.FromString(refreshTokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token id: %w", err)
	}

	refreshToken, err := s.refreshTokens.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return RefreshTokenRequestFromBusiness(refreshToken), nil
}

// TerminateSession implements the op.Storage interface
// it will be called after the user signed out, therefore the access and refresh token of the user of this client must be removed
func (s *Storage) TerminateSession(ctx context.Context, userID string, clientID string) error {
	err := s.refreshTokens.TerminateSessions(ctx, userID, clientID)
	if err != nil {
		return fmt.Errorf("error terminating session: %w", err)
	}

	return nil
}

// RevokeToken implements the op.Storage interface
// it will be called after parsing and validation of the token revocation request
func (s *Storage) RevokeToken(ctx context.Context, tokenOrTokenID string, userID string, clientID string) *oidc.Error {
	uid, err := uuid.FromString(tokenOrTokenID)
	if err != nil {
		return oidc.ErrInvalidRequest().WithDescription("invalid accessToken")
	}

	accessToken, err := s.accessTokens.Get(ctx, uid)
	if err == nil && accessToken != nil {
		if accessToken.ClientID != clientID {
			return oidc.ErrInvalidClient().WithDescription("accessToken was not issued for this client")
		}

		err = s.accessTokens.Delete(ctx, *accessToken)
		if err != nil {
			return oidc.ErrServerError().WithDescription(err.Error())
		}

		return nil
	}

	refreshToken, err := s.refreshTokens.Get(ctx, uid)
	if err == nil && refreshToken == nil {
		// if the token is neither an access nor a refresh token, just ignore it, the expected behaviour of
		// being not valid (anymore) is achieved
		return nil
	}

	if err != nil {
		return oidc.ErrServerError().WithDescription("failed to get refreshToken")
	}

	if accessToken.ClientID != clientID {
		return oidc.ErrInvalidClient().WithDescription("refreshToken was not issued for this client")
	}

	// This should also take care of deleting the access token
	err = s.refreshTokens.Delete(ctx, *refreshToken)
	if err != nil {
		return oidc.ErrServerError().WithDescription(err.Error())
	}

	return nil
}

// GetRefreshTokenInfo looks up a refresh token and returns the token id and user id.
// If given something that is not a refresh token, it must return error.
func (s *Storage) GetRefreshTokenInfo(ctx context.Context, clientID string, tokenStr string) (userID string, tokenID string, err error) {
	uid, err := uuid.FromString(tokenStr)
	if err != nil {
		return "", "", op.ErrInvalidRefreshToken
	}

	token, err := s.refreshTokens.Get(ctx, uid)
	if err == nil && token == nil {
		return "", "", op.ErrInvalidRefreshToken
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	if token.ClientID != clientID {
		return "", "", op.ErrInvalidRefreshToken
	}

	return token.UserID, token.ID.String(), nil
}

// SigningKey implements the op.Storage interface
// it will be called when creating the OpenID Provider
func (s *Storage) SigningKey(ctx context.Context) (op.SigningKey, error) {
	key, err := s.keys.GetSigningKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}

	if key == nil {
		return nil, fmt.Errorf("no signing key found")
	}

	return key, nil
}

// SignatureAlgorithms implements the op.Storage interface
// it will be called to get the sign
func (s *Storage) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	key, err := s.keys.GetSigningKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}

	if key == nil {
		return nil, fmt.Errorf("no signing key found")
	}

	return []jose.SignatureAlgorithm{key.SignatureAlgorithm()}, nil
}

// KeySet implements the op.Storage interface
// it will be called to get the current (public) keys, among others for the keys_endpoint or for validating access_tokens on the userinfo_endpoint, ...
func (s *Storage) KeySet(ctx context.Context) ([]op.Key, error) {
	keys, err := s.keys.GetPublicKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing keys: %w", err)
	}

	var opKeys []op.Key
	for _, key := range keys {
		opKeys = append(opKeys, &key)
	}

	return opKeys, nil
}

// GetClientByClientID implements the op.Storage interface
// it will be called whenever information (type, redirect_uris, ...) about the client behind the client_id is needed
func (s *Storage) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	client, ok := s.clients[clientID]
	if !ok {
		return nil, oidc.ErrInvalidClient()
	}

	return client, nil
}

// AuthorizeClientIDSecret implements the op.Storage interface
// it will be called for validating the client_id, client_secret on token or introspection requests
func (s *Storage) AuthorizeClientIDSecret(ctx context.Context, clientID, clientSecret string) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	client, ok := s.clients[clientID]
	if !ok {
		return oidc.ErrInvalidClient()
	}

	if client.secret != clientSecret {
		return oidc.ErrUnauthorizedClient()
	}

	return nil
}

// SetUserinfoFromScopes implements the op.Storage interface.
// Provide an empty implementation and use SetUserinfoFromRequest instead.
func (s *Storage) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, userID, clientID string, scopes []string) error {
	return nil
}

// setUserinfo sets the info based on the user, scopes and if necessary the clientID
func (s *Storage) setUserinfo(ctx context.Context, userInfo *oidc.UserInfo, userID, clientID string, scopes []string) (err error) {
	uid, err := uuid.FromString(userID)
	if err != nil {
		return fmt.Errorf("invalid userID")
	}

	user, err := s.users.Get(uid)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeOpenID:
			userInfo.Subject = user.ID.String()
		case oidc.ScopeEmail:
			primaryEmail := user.Emails.GetPrimary()
			if primaryEmail == nil {
				return fmt.Errorf("no primary email found")
			}

			userInfo.Email = primaryEmail.Address
			userInfo.EmailVerified = oidc.Bool(primaryEmail.Verified)
			/*
				case oidc.ScopeProfile:
					userInfo.PreferredUsername = user.Username
					userInfo.Name = user.FirstName + " " + user.LastName
					userInfo.FamilyName = user.LastName
					userInfo.GivenName = user.FirstName
					userInfo.Locale = oidc.NewLocale(user.PreferredLanguage)
				case oidc.ScopePhone:
					userInfo.PhoneNumber = user.Phone
					userInfo.PhoneNumberVerified = user.PhoneVerified
			*/
		}
	}
	return nil
}

// SetUserinfoFromToken implements the op.Storage interface
// it will be called for the userinfo endpoint, so we read the token and pass the information from that to the private function
func (s *Storage) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID, subject, origin string) error {
	uid, err := uuid.FromString(tokenID)
	if err != nil {
		return fmt.Errorf("failed to parse token id: %w", err)
	}

	token, err := s.accessTokens.Get(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if token == nil {
		return fmt.Errorf("token not found")
	}

	if token.Expiration.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	return s.setUserinfo(ctx, userinfo, token.Subject, token.ClientID, token.Scopes)
}

// SetIntrospectionFromToken implements the op.Storage interface
// it will be called for the introspection endpoint, so we read the token and pass the information from that to the private function
func (s *Storage) SetIntrospectionFromToken(ctx context.Context, userinfo *oidc.IntrospectionResponse, tokenID, subject, clientID string) error {
	uid, err := uuid.FromString(tokenID)
	if err != nil {
		return fmt.Errorf("failed to parse token id: %w", err)
	}

	token, err := s.accessTokens.Get(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if token == nil {
		return fmt.Errorf("token not found")
	}

	if token.Expiration.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	for _, aud := range token.Audience {
		if aud == clientID {
			// the introspection response only has to return a boolean (active) if the token is active
			// this will automatically be done by the library if you don't return an error
			// you can also return further information about the user / associated token
			// e.g. the userinfo (equivalent to userinfo endpoint)

			userInfo := new(oidc.UserInfo)
			err := s.setUserinfo(ctx, userInfo, subject, clientID, token.Scopes)
			if err != nil {
				return err
			}

			userinfo.SetUserInfo(userInfo)
			//...and also the requested scopes...
			userinfo.Scope = token.Scopes
			//...and the client the token was issued to
			userinfo.ClientID = token.ClientID

			return nil
		}
	}

	return fmt.Errorf("token is not valid for this client")
}

func (s *Storage) getPrivateClaimsFromScopes(ctx context.Context, userID, clientID string, scopes []string) (claims map[string]interface{}, err error) {
	for _, scope := range scopes {
		switch scope {
		}
	}
	return claims, nil
}

// GetPrivateClaimsFromScopes implements the op.Storage interface
// it will be called for the creation of a JWT access token to assert claims for custom scopes
func (s *Storage) GetPrivateClaimsFromScopes(ctx context.Context, userID, clientID string, scopes []string) (map[string]interface{}, error) {
	return s.getPrivateClaimsFromScopes(ctx, userID, clientID, scopes)
}

// GetKeyByIDAndClientID implements the op.Storage interface
// it will be called to validate the signatures of a JWT (JWT Profile Grant and Authentication)
func (s *Storage) GetKeyByIDAndClientID(ctx context.Context, keyID, clientID string) (*jose.JSONWebKey, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	client, ok := s.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("clientID not found")
	}

	key, ok := client.GetKey(keyID)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	return &jose.JSONWebKey{
		KeyID: keyID,
		Use:   "sig",
		Key:   key,
	}, nil
}

// ValidateJWTProfileScopes implements the op.Storage interface
// it will be called to validate the scopes of a JWT Profile Authorization Grant request
func (s *Storage) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
	allowedScopes := make([]string, 0)
	for _, scope := range scopes {
		if scope == oidc.ScopeOpenID {
			allowedScopes = append(allowedScopes, scope)
		}
	}
	return allowedScopes, nil
}

func (s *Storage) Health(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

// SetUserinfoFromRequest implements the op.CanSetUserinfoFromRequest interface.  In the
// next major release, it will be required for op.Storage.
// It will be called for the creation of an id_token, so we'll just pass it to the private function without any further check
func (s *Storage) SetUserinfoFromRequest(ctx context.Context, userinfo *oidc.UserInfo, token op.IDTokenRequest, scopes []string) error {
	return s.setUserinfo(ctx, userinfo, token.GetSubject(), token.GetClientID(), scopes)
}

// getInfoFromRequest returns the clientID, authTime and amr depending on the op.TokenRequest type / implementation
func getInfoFromRequest(req op.TokenRequest) (clientID string, authTime time.Time, amr []string) {
	authReq, ok := req.(*AuthRequest) // Code Flow (with scope offline_access)
	if ok {
		return authReq.ApplicationID, authReq.authTime, authReq.GetAMR()
	}

	refreshReq, ok := req.(*RefreshTokenRequest) // Refresh Token Request
	if ok {
		return refreshReq.ClientID, refreshReq.AuthTime, refreshReq.AMR
	}

	return "", time.Time{}, nil
}
