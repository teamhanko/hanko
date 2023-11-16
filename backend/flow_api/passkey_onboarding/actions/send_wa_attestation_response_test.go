package actions

import (
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	capabilitiesStates "github.com/teamhanko/hanko/backend/flow_api/capabilities/states"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"testing"
	"time"
)

func TestSendWaAttestationResponse(t *testing.T) {
	s := new(sendWaAttestationResponse)

	suite.Run(t, s)
}

type sendWaAttestationResponse struct {
	test.Suite
}

func (s *sendWaAttestationResponse) TestSendWaAttestationResponse_Execute() {
	tests := []struct {
		name          string
		input         string
		flowId        string
		expectedState flowpilot.StateName
		statusCode    int
	}{
		{
			name:          "send a correct attestation",
			input:         `{"public_key": "{\"id\": \"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH\",\"rawId\": \"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH\",\"type\": \"public-key\",\"response\": {\"attestationObject\": \"o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVjeSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFYmehnq3OAAI1vMYKZIsLJfHwVQMAWgGhXZHA-Erj4xfo8FKEcB_PmR7mOUVuOn7GZhLwV-kTSh2hrVc6QE7NOikFYXiDo2M_mJ3huHJkDnnc5dHtIxfedbpMdex5fY3hoFs-fwymQjtdqdvti5c4x6UBAgMmIAEhWCDxvVrRgK4vpnr6JxTx-KfpSNyQUtvc47ryryZmj-P5kSJYIDox8N9bHQBrxN-b5kXqfmj3GwAJW7nNCh8UPbus3B6I\",\"clientDataJSON\": \"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoidE9yTkRDRDJ4UWY0ekZqRWp3eGFQOGZPRXJQM3p6MDhyTW9UbEpHdG5LVSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0\"}}"}`,
			flowId:        "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			expectedState: shared.StateSuccess,
			statusCode:    http.StatusOK,
		},
		{
			name:          "send a attestation with expired session data",
			input:         `{"public_key": "{\"id\": \"4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM\",\"rawId\": \"4iVZGFN_jktXJmwmBmaSq0Qr4T62T0jX7PS7XcgAWlM\",\"type\": \"public-key\",\"response\": {\"attestationObject\": \"o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVikSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFAAAAAQECAwQFBgcIAQIDBAUGBwgAIOIlWRhTf45LVyZsJgZmkqtEK-E-tk9I1-z0u13IAFpTpQECAyYgASFYIAeA_nt5TQ8c7bc8hN9_3zqzp3coXO5aplEeHMOQG0hrIlggf_KVxZI_nIedc1XMrwwOMaYNd0qxVpFK7vU79fGBoxY\",\"clientDataJSON\": \"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiRmVNYzdzUjlFbGVod0VVNVR0RVdGaTdyUFAzLWtkWlhnbndMdGxiM0NoWSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODg4OCIsImNyb3NzT3JpZ2luIjpmYWxzZX0\"}}"}`,
			flowId:        "53d35f35-c87d-4533-b966-2b48686b9be9",
			expectedState: states.StateOnboardingVerifyPasskeyAttestation,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:          "send a attestation with wrong challenge",
			input:         `{"publicKey":"{\"id\": \"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH\",\"rawId\": \"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH\",\"type\": \"public-key\",\"response\": {\"attestationObject\": \"o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVjeSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFYmehnq3OAAI1vMYKZIsLJfHwVQMAWgGhXZHA-Erj4xfo8FKEcB_PmR7mOUVuOn7GZhLwV-kTSh2hrVc6QE7NOikFYXiDo2M_mJ3huHJkDnnc5dHtIxfedbpMdex5fY3hoFs-fwymQjtdqdvti5c4x6UBAgMmIAEhWCDxvVrRgK4vpnr6JxTx-KfpSNyQUtvc47ryryZmj-P5kSJYIDox8N9bHQBrxN-b5kXqfmj3GwAJW7nNCh8UPbus3B6I\",\"clientDataJSON\": \"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoidE9yTkRDRDJ4UWY0ekZqRWp3eGFQOGZPRXJQM3p6MDhyTW9UbEpHdG5LdSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0\"}}"}`,
			flowId:        "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			expectedState: states.StateOnboardingVerifyPasskeyAttestation,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:          "error state and forbidden status if webauthn not available and action is suspended",
			input:         `{"publicKey":"{\"id\": \"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH\",\"rawId\": \"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH\",\"type\": \"public-key\",\"response\": {\"attestationObject\": \"o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVjeSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFYmehnq3OAAI1vMYKZIsLJfHwVQMAWgGhXZHA-Erj4xfo8FKEcB_PmR7mOUVuOn7GZhLwV-kTSh2hrVc6QE7NOikFYXiDo2M_mJ3huHJkDnnc5dHtIxfedbpMdex5fY3hoFs-fwymQjtdqdvti5c4x6UBAgMmIAEhWCDxvVrRgK4vpnr6JxTx-KfpSNyQUtvc47ryryZmj-P5kSJYIDox8N9bHQBrxN-b5kXqfmj3GwAJW7nNCh8UPbus3B6I\",\"clientDataJSON\": \"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoidE9yTkRDRDJ4UWY0ekZqRWp3eGFQOGZPRXJQM3p6MDhyTW9UbEpHdG5LdSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0\"}}"}`,
			flowId:        "be57518c-6bd5-4b3e-a91a-6c082e212a58",
			expectedState: shared.StateError,
			statusCode:    http.StatusForbidden,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../../test/fixtures/actions/send_wa_attestation_response")
			s.Require().NoError(err)

			wa, err := s.getWebauthnLib()
			s.Require().NoError(err)

			passkeySubFlow, err := flowpilot.NewSubFlow().
				State(states.StateOnboardingCreatePasskey).
				State(states.StateOnboardingVerifyPasskeyAttestation, NewSendWAAttestationResponse(s.Storage, wa)).
				Build()
			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(capabilitiesStates.StatePreflight).
				State(shared.StateSuccess).
				SubFlows(passkeySubFlow).
				InitialState(capabilitiesStates.StatePreflight).
				ErrorState(shared.StateError).
				Debug(true).
				Build()
			s.Require().NoError(err)

			tx := s.Storage.GetConnection()
			db := models.NewFlowDB(tx)
			actionParam := fmt.Sprintf("send_wa_attestation_response@%s", currentTest.flowId)
			inputData := flowpilot.InputData{JSONString: currentTest.input}
			result, err := flow.Execute(db, flowpilot.WithActionParam(actionParam), flowpilot.WithInputData(inputData))
			s.Require().NoError(err)

			s.Equal(currentTest.statusCode, result.Status())
			s.Equal(currentTest.expectedState, result.Response().StateName)

			// TODO: check that the schema of the action returns the correct error_code e.g.
			// result.Response().PublicActions[0].PublicSchema[0].PublicError.Code == ErrorValueInvalid
		})
	}
}

func (s *sendWaAttestationResponse) getWebauthnLib() (*webauthnLib.WebAuthn, error) {
	f := false
	return webauthnLib.New(&webauthnLib.Config{
		RPID:                  "localhost",
		RPDisplayName:         "Test RP",
		RPOrigins:             []string{"http://localhost:8080"},
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &f,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Debug: false,
		Timeouts: webauthnLib.TimeoutsConfig{
			Login: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: 60000 * time.Millisecond,
			},
			Registration: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: 60000 * time.Millisecond,
			},
		},
	})
}

func (s *sendWaAttestationResponse) GetConfig() config.Config {
	return config.Config{
		Webauthn: config.WebauthnSettings{
			RelyingParty: config.RelyingParty{
				Id:          "localhost",
				DisplayName: "Test Relying Party",
				Icon:        "",
				Origins:     []string{"http://localhost:8080"},
			},
			Timeout: 60000,
		},
	}
}
