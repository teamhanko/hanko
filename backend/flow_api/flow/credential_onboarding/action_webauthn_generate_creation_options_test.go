package credential_onboarding

import (
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/flow/capabilities"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"testing"
	"time"
)

func TestGetWaCreationOptions(t *testing.T) {
	s := new(getWaCreationOptions)

	suite.Run(t, s)
}

type getWaCreationOptions struct {
	test.Suite
}

func (s *getWaCreationOptions) TestGetWaCreationOptions_Execute() {
	tests := []struct {
		name          string
		input         string
		flowId        string
		cfg           config.Config
		expectedState flowpilot.StateName
		statusCode    int
	}{
		{
			name:          "get webauthn creation options with username and email",
			input:         "",
			flowId:        "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg:           config.Config{},
			expectedState: StateOnboardingVerifyPasskeyAttestation,
			statusCode:    http.StatusOK,
		},
		{
			name:          "get webauthn creation options with only email",
			input:         "",
			flowId:        "a77e23b2-7ca5-4c76-a20b-c17b7dbcb117",
			cfg:           config.Config{},
			expectedState: StateOnboardingVerifyPasskeyAttestation,
			statusCode:    http.StatusOK,
		},
		{
			name:          "get webauthn creation options with only username",
			input:         "",
			flowId:        "de87cfc6-a6e2-434d-bbe8-5e5004c9deda",
			cfg:           config.Config{},
			expectedState: StateOnboardingVerifyPasskeyAttestation,
			statusCode:    http.StatusOK,
		},
		{
			name:          "error state and forbidden status if webauthn not available and action is suspended",
			input:         "",
			flowId:        "be57518c-6bd5-4b3e-a91a-6c082e212a58",
			cfg:           config.Config{},
			expectedState: shared.StateError,
			statusCode:    http.StatusForbidden,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../../test/fixtures/actions/get_wa_creation_options")
			s.Require().NoError(err)

			wa, err := s.getWebauthnLib()
			s.Require().NoError(err)

			passkeySubFlow, err := flowpilot.NewSubFlow().
				State(StateOnboardingCreatePasskey, WebauthnGenerateCreationOptions{}).
				State(StateOnboardingVerifyPasskeyAttestation).
				Build()
			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(capabilities.StatePreflight).
				State(shared.StateSuccess).
				SubFlows(passkeySubFlow).
				InitialState(capabilities.StatePreflight).
				ErrorState(shared.StateError).
				Build()
			s.Require().NoError(err)

			tx := s.Storage.GetConnection()
			db := models.NewFlowDB(tx)
			actionParam := fmt.Sprintf("get_wa_creation_options@%s", currentTest.flowId)
			inputData := flowpilot.InputData{JSONString: currentTest.input}
			result, err := flow.Execute(db, flowpilot.WithQueryParam(actionParam), flowpilot.WithInputData(inputData))
			s.Require().NoError(err)

			s.Equal(currentTest.statusCode, result.Status())
			s.Equal(currentTest.expectedState, result.Response().Name)

			// TODO: check that the schema of the action returns the correct error_code e.g.
			// result.Response().ResponseActions[0].ResponseInputs[0].ResponseError.Code == ErrorValueInvalid
		})
	}
}

func (s *getWaCreationOptions) getWebauthnLib() (*webauthnLib.WebAuthn, error) {
	f := false
	return webauthnLib.New(&webauthnLib.Config{
		RPID:                  "localhost",
		RPDisplayName:         "Test RP",
		RPOrigins:             []string{"http://localhost"},
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
