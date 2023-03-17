package thirdparty

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/crypto/aes_gcm"
	"time"
)

func GenerateState(config *config.Config, provider string, redirectTo string) ([]byte, error) {
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
	Provider   string    `json:"provider"`
	RedirectTo string    `json:"redirect_to"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Nonce      string    `json:"nonce"`
}

func VerifyState(config *config.Config, state string, expectedState string) (*State, error) {
	decodedState, err := decodeState(config, state)
	if err != nil {
		return nil, fmt.Errorf("could not decode state: %w", err)
	}

	decodedExpectedState, err := decodeState(config, expectedState)
	if err != nil {
		return nil, fmt.Errorf("could not decode expectedState: %w", err)
	}

	if decodedState.Nonce != decodedExpectedState.Nonce {
		return nil, errors.New("could not verify state")
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
