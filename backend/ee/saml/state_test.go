package saml

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"strings"
	"testing"
	"time"
)

func TestSaml_GenerateState(t *testing.T) {
	cfg := &config.Config{
		Secrets: config.Secrets{Keys: []string{"thirty-two-byte-long-test-secret"}},
		Saml: samlConfig.Saml{
			DefaultRedirectUrl: "https://example.com",
		},
	}

	persister := test.NewSamlStatePersister([]models.SamlState{})

	state, err := GenerateState(cfg, persister, "test-provider", "https://example.com")
	assert.NoError(t, err)
	assert.NotNil(t, state)
}

func TestSaml_GenerateStateWithDefaultRedirect(t *testing.T) {
	cfg := &config.Config{
		Secrets: config.Secrets{Keys: []string{"thirty-two-byte-long-test-secret"}},
		Saml: samlConfig.Saml{
			DefaultRedirectUrl: "https://example.com",
		},
	}

	persister := test.NewSamlStatePersister([]models.SamlState{})

	state, err := GenerateState(cfg, persister, "test-provider", "")
	assert.NoError(t, err)
	assert.NotNil(t, state)
}

func TestSaml_GenerateState_Error(t *testing.T) {

	redirectTo := "https://example.com"

	tests := []struct {
		name             string
		provider         string
		redirectTo       string
		secret           string
		errorRedirectUrl string
		expectedError    string
	}{
		{
			name:             "provider is not present",
			secret:           "test-secret",
			provider:         "",
			redirectTo:       redirectTo,
			errorRedirectUrl: redirectTo,
			expectedError:    "provider must be present",
		},
		{
			name:             "provider is not present",
			secret:           "",
			provider:         "test-provider",
			redirectTo:       redirectTo,
			errorRedirectUrl: redirectTo,
			expectedError:    "could not instantiate aesgcm",
		},
	}
	for _, testData := range tests {
		t.Run(testData.name, func(t *testing.T) {
			cfg := &config.Config{
				Secrets: config.Secrets{Keys: []string{testData.secret}},
			}
			persister := test.NewSamlStatePersister([]models.SamlState{})

			_, err := GenerateState(cfg, persister, testData.provider, testData.redirectTo)
			assert.NotNil(t, err)
			assert.True(t, strings.Contains(err.Error(), testData.expectedError))
		})
	}
}

func TestSaml_VerifyState(t *testing.T) {
	cfg := &config.Config{
		Secrets: config.Secrets{Keys: []string{"thirty-two-byte-long-test-secret"}},
	}

	state := "HmD7wlGQ7bF_4MGtmFRQuuSGTshHETDs4RQa64JAx-6EsmNsUjaQwYNOnjWUs6qIOuQMBTKapDGVXVCk00pX2vSS-x-WVqdzZ8KyeQ-9IHu2mwb-AeRbb2QPE-GFnvp2wrbCskKvWvtOfipyeTsnYY5iM90DxssaUtvKnawaB5_MNNekfKyiOeepIkKjUfSJ6-yTR7AAA4B9jwOfDRB4zdV8kKPVJlGVBJFosL11YWJaLxRGQR69nah3Jf9Z6bSAGXxWp24PoBYhij-dH4JyDCcU7D-NeT2A8qFFFjQ1m28C8fsr6zqb4w=="

	persister := test.NewSamlStatePersister([]models.SamlState{
		{
			ID:        uuid.UUID{},
			Nonce:     "VSPVgv0ZWxnvZtVkh7imAu1KQRONFsUhGDxgp9x82aU=",
			State:     state,
			ExpiresAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	})

	validState, err := VerifyState(cfg, persister, state)
	require.NoError(t, err)
	require.NotNil(t, validState)
	require.Equal(t, "test-provider", validState.Provider)
	require.Equal(t, "https://example.com", validState.RedirectTo)
	require.NotEmpty(t, validState.Nonce)
	require.True(t, time.Now().UTC().Before(validState.ExpiresAt))
}
func TestSaml_VerifyState_Error(t *testing.T) {
	tests := []struct {
		name          string
		nonce         string
		state         string
		expectedState string
		expectedError string
	}{
		{
			name:          "error on invalid state",
			nonce:         "lorem",
			state:         "invalid_state",
			expectedError: "could not decode state",
		},
		{
			name:          "error on invalid expected state",
			nonce:         "AxFqBzDKttD_v-m-GxrTSXJbYKFVnn7eloTLzoyYwJ0=",
			state:         "ctj9hAU6kFkdc-g5ZXWGHbbnE1NSoQ55afdcYWLImTc5C5gGdeaknPDvZ560LJ07uA8I7X2ssPBKkkQb5xDnDHykmQtEp1hBp0uP_PdQdNJZ7FVXb0MpVWMrCv9nZp_fquLQvzjdm2nGRP9VKPt6S_7XNqg_mMA_ri9g7XWuNgJUqUE0PW7TGB9otsD2E7FQNowAl_vY-q0mYYIdkl0qb3lxWxnSzP51ewVv5VTeI7Mfr1gXyxWyTdqTnl_s5azaRq6w8J2SSX1_GTppX49zqCidPVIqYp6IyvotyOz-ePmvTg7tizBsaQ==",
			expectedState: "invalid_state",
			expectedError: "could not decode expectedState",
		},
		{
			name:          "error on nonce mismatch",
			nonce:         "AxFqBzDKttD_v-m-GxrTSXJbYKFVnn7eloTLzoyYwJ0=",
			state:         "ctj9hAU6kFkdc-g5ZXWGHbbnE1NSoQ55afdcYWLImTc5C5gGdeaknPDvZ560LJ07uA8I7X2ssPBKkkQb5xDnDHykmQtEp1hBp0uP_PdQdNJZ7FVXb0MpVWMrCv9nZp_fquLQvzjdm2nGRP9VKPt6S_7XNqg_mMA_ri9g7XWuNgJUqUE0PW7TGB9otsD2E7FQNowAl_vY-q0mYYIdkl0qb3lxWxnSzP51ewVv5VTeI7Mfr1gXyxWyTdqTnl_s5azaRq6w8J2SSX1_GTppX49zqCidPVIqYp6IyvotyOz-ePmvTg7tizBsaQ==",
			expectedState: "Fs1SEytL1YuZJvBCyfF3Iydpl3-aC95bDsIb3VIcbq3sZjhn6iQaNFVNQZG1Maz2bI7Zkz3UvEIrIFohIotDqoR057mBnDKHVeg-TZKLXfviPJqfkOldfNPTpJgO9e4biYGxPx6qkmabk83eO77Qe3rJ2XM6FznQqPoMjq1vOBBJXjBUPeu3KtF1l7ONpHHVCV1Sr0cm0qQ4q5nPYgjI227AOa_MOOIDKfwqxb-jQT_9n0vLtVen9aYFbr_i_57r5aC3nCxnBgq7gXDq7z-E5NDRW23E0x-5CpQdDhElmtoeu4XRx1jOfw==",
			expectedError: "could not verify state",
		},
		{
			name:          "error on expired state",
			nonce:         "lmBPwLelStCMc--HmSVb-VaEDIHXVSY0d81EHMpI9nc=",
			state:         "UHPShpOq0byNI-A1vvUROLsPjNhGF5Xxdme4llZrnCvXfyZP_RnhRq490XPqKGKeWa621MtAUsV7N6C4OGx-EXK1TaLWNIYfFmByIEIlPSkpMVTFEUedY5UaFXwhWhuQb4Ci0r3QlPCLmQnPin1O4Vb59K2KieJwTvVnZzY3Y3Avj-D91acTMRGN6OabuIDDOH4nl0qDJABkZ6tnYk735ot8s3oyvJmZgmsW0qLOC_OoNMGZqLUTQzCrEayJ-gTXKr6HSvClN-4KGi4htHHLwYjrSCr_6tnQxDhZvNq7GKgXUEXfeUmZGQ==",
			expectedState: "UHPShpOq0byNI-A1vvUROLsPjNhGF5Xxdme4llZrnCvXfyZP_RnhRq490XPqKGKeWa621MtAUsV7N6C4OGx-EXK1TaLWNIYfFmByIEIlPSkpMVTFEUedY5UaFXwhWhuQb4Ci0r3QlPCLmQnPin1O4Vb59K2KieJwTvVnZzY3Y3Avj-D91acTMRGN6OabuIDDOH4nl0qDJABkZ6tnYk735ot8s3oyvJmZgmsW0qLOC_OoNMGZqLUTQzCrEayJ-gTXKr6HSvClN-4KGi4htHHLwYjrSCr_6tnQxDhZvNq7GKgXUEXfeUmZGQ==",
			expectedError: "state is expired",
		},
	}
	for _, testData := range tests {
		t.Run(testData.name, func(t *testing.T) {
			cfg := &config.Config{
				Secrets: config.Secrets{Keys: []string{"thirty-two-byte-long-test-secret"}},
			}

			persister := test.NewSamlStatePersister([]models.SamlState{
				{
					ID:        uuid.UUID{},
					Nonce:     testData.nonce,
					State:     testData.expectedState,
					ExpiresAt: time.Now(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			})

			_, err := VerifyState(cfg, persister, testData.state)
			assert.NotNil(t, err)
			assert.True(t, strings.Contains(err.Error(), testData.expectedError))
		})
	}
}
