package main

import (
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/zitadel/oidc/v2/pkg/op"
	"time"
)

// RefreshTokenRequestFromBusiness will simply wrap the storage RefreshToken to implement the op.RefreshTokenRequest interface
func RefreshTokenRequestFromBusiness(token *models.RefreshToken) op.RefreshTokenRequest {
	return &RefreshTokenRequest{token}
}

type RefreshTokenRequest struct {
	*models.RefreshToken
}

func (r *RefreshTokenRequest) GetAMR() []string {
	return r.AMR
}

func (r *RefreshTokenRequest) GetAudience() []string {
	return r.Audience
}

func (r *RefreshTokenRequest) GetAuthTime() time.Time {
	return r.AuthTime
}

func (r *RefreshTokenRequest) GetClientID() string {
	return r.ClientID
}

func (r *RefreshTokenRequest) GetScopes() []string {
	return r.Scopes
}

func (r *RefreshTokenRequest) GetSubject() string {
	return r.UserID
}

func (r *RefreshTokenRequest) SetCurrentScopes(scopes []string) {
	r.Scopes = scopes
}
