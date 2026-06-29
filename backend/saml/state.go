package saml

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto"
	"github.com/teamhanko/hanko/backend/v3/crypto/aes_gcm"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type State struct {
	Provider   string    `json:"provider"`
	RedirectTo string    `json:"redirect_to"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Nonce      string    `json:"nonce"`
	IsFlow     bool      `json:"is_flow"`
}

const StatePrefixServiceProviderInitiated = "hanko_spi_"

func GenerateStateForFlowAPI(isFlow bool) func(*State) {
	return func(state *State) {
		state.IsFlow = isFlow
	}
}

func GenerateState(config config.Config, persister persistence.SamlStatePersister, provider string, redirectTo string, tenantID uuid.UUID, options ...func(*State)) ([]byte, error) {
	if strings.TrimSpace(provider) == "" {
		return nil, errors.New("provider must be present")
	}

	if strings.TrimSpace(redirectTo) == "" {
		redirectTo = config.TenantConfig.Saml.DefaultRedirectUrl
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

	aes, err := aes_gcm.NewAESGCM(config.ApplicationConfig.SecretKeys)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate aesgcm: %w", err)
	}

	encryptedState, err := aes.Encrypt(stateJson)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt state: %w", err)
	}

	dbState, err := models.NewSamlState(nonce, encryptedState, tenantID)
	if err != nil {
		return nil, fmt.Errorf("could not create state model: %w", err)
	}

	err = persister.Create(*dbState)
	if err != nil {
		return nil, fmt.Errorf("could not save state to db: %w", err)
	}

	// Add prefix to distinguish between SP initiated and IDP initiated requests in callback handler.
	result := fmt.Sprintf("%s%s", StatePrefixServiceProviderInitiated, encryptedState)
	return []byte(result), nil
}

func VerifyState(tenantID uuid.UUID, keys []string, persister persistence.SamlStatePersister, state string) (*State, error) {
	decodedState, err := decodeState(keys, state)
	if err != nil {
		return nil, fmt.Errorf("could not decode state: %w", err)
	}

	expectedState, err := persister.GetByNonce(decodedState.Nonce, tenantID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch expected state from db: %w", err)
	}

	decodedExpectedState, err := decodeState(keys, expectedState.State)
	if err != nil {
		return nil, fmt.Errorf("could not decode expectedState: %w", err)
	}

	if decodedState.Nonce != decodedExpectedState.Nonce {
		return nil, errors.New("could not verify state")
	}

	_ = persister.Delete(*expectedState)

	if time.Now().UTC().After(decodedState.ExpiresAt) {
		return nil, errors.New("state is expired")
	}

	return decodedState, nil
}

func decodeState(keys []string, state string) (*State, error) {
	aes, err := aes_gcm.NewAESGCM(keys)
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
