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
		Expiration: now.Add(time.Minute * 5),
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
	Expiration time.Time `json:"expiration"`
	Nonce      string    `json:"nonce"`
}

func VerifyState(config *config.Config, state string) (*State, error) {
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

	if time.Now().UTC().After(unmarshalledState.Expiration) {
		return nil, errors.New("state is expired")
	}

	return &unmarshalledState, nil
}
