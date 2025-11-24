package thirdparty

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto"
	"github.com/teamhanko/hanko/backend/v2/crypto/aes_gcm"
)

func GenerateStateForFlowAPI(isFlow bool) func(*State) {
	return func(state *State) {
		state.IsFlow = isFlow
	}
}

func GenerateStateWithPKCECodeVerifier(codeVerifier string) func(state *State) {
	return func(state *State) {
		if codeVerifier != "" {
			state.CodeVerifier = codeVerifier
		}
	}
}

// GenerateStateForLoggedInUser If the state is generated for a logged-in user, the OAuth request and response must only be used with the same already logged-in user.
func GenerateStateForLoggedInUser(userID uuid.UUID) func(*State) {
	return func(state *State) {
		if userID != uuid.Nil {
			state.UserID = &userID
		}
	}
}

func GenerateState(config *config.Config, provider string, redirectTo string, options ...func(*State)) ([]byte, error) {
	if provider == "" {
		return nil, errors.New("provider must be present")
	}

	if redirectTo == "" {
		redirectTo = config.ThirdParty.ErrorRedirectURL
	}

	nonce, err := crypto.GenerateRandomStringURLSafe(32)
	if err != nil {
		return nil, fmt.Errorf("could not generate nonce: %w", err)
	}

	now := time.Now().UTC()
	state := State{
		Provider:   provider,
		RedirectTo: redirectTo,
		IssuedAt:   now,
		ExpiresAt:  now.Add(time.Minute * 5),
		Nonce:      nonce,
	}

	for _, option := range options {
		option(&state)
	}

	stateJson, err := json.Marshal(state)

	aes, err := aes_gcm.NewAESGCM(config.Secrets.Keys)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate aesgcm: %w", err)
	}

	encryptedState, err := aes.Encrypt(stateJson)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt state: %w", err)
	}

	return []byte(encryptedState), nil
}

type State struct {
	Provider     string     `json:"provider"`
	RedirectTo   string     `json:"redirect_to"`
	IssuedAt     time.Time  `json:"issued_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	Nonce        string     `json:"nonce"`
	IsFlow       bool       `json:"is_flow"`
	CodeVerifier string     `json:"code_verifier,omitempty"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
}

func VerifyState(config *config.Config, state string, expectedState string) (*State, error) {
	decodedState, err := decodeState(config, state)
	if err != nil {
		return nil, fmt.Errorf("could not decode state: %w", err)
	}

	if decodedState.CodeVerifier == "" {
		if expectedState == "" {
			return nil, errors.New("expected state must not be empty")
		}
		decodedExpectedState, err := decodeState(config, expectedState)
		if err != nil {
			return nil, fmt.Errorf("could not decode expectedState: %w", err)
		}

		if decodedState.Nonce != decodedExpectedState.Nonce {
			return nil, errors.New("could not verify state")
		}
	}

	if time.Now().UTC().After(decodedState.ExpiresAt) {
		return nil, errors.New("state is expired")
	}

	return decodedState, nil
}

func decodeState(config *config.Config, state string) (*State, error) {
	aes, err := aes_gcm.NewAESGCM(config.Secrets.Keys)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate aesgcm: %w", err)
	}

	decryptedState, err := aes.Decrypt(state)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt state: %w", err)
	}

	var unmarshalledState State
	err = json.Unmarshal(decryptedState, &unmarshalledState)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal state: %w", err)
	}
	return &unmarshalledState, nil
}
