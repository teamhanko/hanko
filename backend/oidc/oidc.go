package main

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/text/language"
	"time"
)

type AuthRequest struct {
	ID            uuid.UUID
	CreationDate  time.Time
	ApplicationID string
	CallbackURI   string
	TransferState string
	Prompt        []string
	UiLocales     []language.Tag
	LoginHint     string
	MaxAuthAge    *time.Duration
	UserID        string
	Scopes        []string
	ResponseType  oidc.ResponseType
	Nonce         string
	CodeChallenge string

	done     bool
	authTime time.Time
}

func NewAuthRequestFromModel(request *models.AuthRequest) (*AuthRequest, error) {
	if request == nil {
		return nil, errors.New("auth request not found")
	}

	var uiLocales []language.Tag
	for _, tag := range request.UILocales {
		uiLocales = append(uiLocales, language.Make(tag))
	}

	return &AuthRequest{
		ID:            request.ID,
		CreationDate:  request.CreatedAt,
		ApplicationID: request.ClientID,
		CallbackURI:   request.CallbackURI,
		TransferState: request.TransferState,
		Prompt:        request.Prompt,
		UiLocales:     uiLocales,
		LoginHint:     request.LoginHint,
		MaxAuthAge:    &request.MaxAuthAge,
		UserID:        request.UserID,
		Scopes:        request.Scopes,
		ResponseType:  oidc.ResponseType(request.ResponseType),
		Nonce:         request.Nonce,
		CodeChallenge: request.CodeChallenge,
	}, nil
}

func (a *AuthRequest) GetID() string {
	return a.ID.String()
}

func (a *AuthRequest) GetACR() string {
	return "" // we won't handle acr
}

func (a *AuthRequest) GetAMR() []string {
	// TODO: https://www.rfc-editor.org/rfc/rfc8176.html

	// this example only uses password for authentication
	if a.done {
		return []string{"pwd"}
	}
	return nil
}

func (a *AuthRequest) GetAudience() []string {
	return []string{a.ApplicationID} // this example will always just use the client_id as audience
}

func (a *AuthRequest) GetAuthTime() time.Time {
	return a.authTime
}

func (a *AuthRequest) GetClientID() string {
	return a.ApplicationID
}

func (a *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge {
	return &oidc.CodeChallenge{
		Challenge: a.CodeChallenge,
		Method:    oidc.CodeChallengeMethodS256,
	}
}

func (a *AuthRequest) GetNonce() string {
	return a.Nonce
}

func (a *AuthRequest) GetRedirectURI() string {
	return a.CallbackURI
}

func (a *AuthRequest) GetResponseType() oidc.ResponseType {
	return a.ResponseType
}

func (a *AuthRequest) GetResponseMode() oidc.ResponseMode {
	return "" // we won't handle response mode
}

func (a *AuthRequest) GetScopes() []string {
	return a.Scopes
}

func (a *AuthRequest) GetState() string {
	return a.TransferState
}

func (a *AuthRequest) GetSubject() string {
	return a.UserID
}

func (a *AuthRequest) Done() bool {
	return a.done
}

func (a *AuthRequest) ToModel() models.AuthRequest {
	var locales []string
	for _, locale := range a.UiLocales {
		locales = append(locales, locale.String())
	}

	var maxAuthAge time.Duration
	if a.MaxAuthAge != nil {
		maxAuthAge = *a.MaxAuthAge
	}

	return models.AuthRequest{
		ID:            a.ID,
		CreatedAt:     a.CreationDate,
		ClientID:      a.ApplicationID,
		CallbackURI:   a.CallbackURI,
		TransferState: a.TransferState,
		Prompt:        a.Prompt,
		UILocales:     locales,
		LoginHint:     a.LoginHint,
		MaxAuthAge:    maxAuthAge,
		UserID:        a.UserID,
		Scopes:        a.Scopes,
		ResponseType:  string(a.ResponseType),
		Nonce:         a.Nonce,
		CodeChallenge: a.CodeChallenge,
	}
}

func PromptToInternal(oidcPrompt oidc.SpaceDelimitedArray) []string {
	prompts := make([]string, len(oidcPrompt))
	for _, oidcPrompt := range oidcPrompt {
		switch oidcPrompt {
		case oidc.PromptNone,
			oidc.PromptLogin,
			oidc.PromptConsent,
			oidc.PromptSelectAccount:
			prompts = append(prompts, oidcPrompt)
		}
	}

	return prompts
}

func MaxAgeToInternal(maxAge *uint) *time.Duration {
	if maxAge == nil {
		return nil
	}

	dur := time.Duration(*maxAge) * time.Second
	return &dur
}

func authRequestToInternal(authReq *oidc.AuthRequest, userID string) *AuthRequest {
	return &AuthRequest{
		CreationDate:  time.Now(),
		ApplicationID: authReq.ClientID,
		CallbackURI:   authReq.RedirectURI,
		TransferState: authReq.State,
		Prompt:        PromptToInternal(authReq.Prompt),
		UiLocales:     authReq.UILocales,
		LoginHint:     authReq.LoginHint,
		MaxAuthAge:    MaxAgeToInternal(authReq.MaxAge),
		UserID:        userID,
		Scopes:        authReq.Scopes,
		ResponseType:  authReq.ResponseType,
		Nonce:         authReq.Nonce,
		CodeChallenge: authReq.CodeChallenge,
	}
}
