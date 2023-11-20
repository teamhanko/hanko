package passcode

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	passkeyOnboarding "github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"testing"
)

func TestSubmitPasscode(t *testing.T) {
	s := new(submitPasscodeActionSuite)

	suite.Run(t, s)
}

type submitPasscodeActionSuite struct {
	test.Suite
}

func (s *submitPasscodeActionSuite) TestSubmitPasscode_Execute() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	tests := []struct {
		name             string
		input            string
		cfg              config.Config
		expectedState    flowpilot.StateName
		statusCode       int
		flowId           string
		expectsFlowError bool
	}{
		{
			name:   "submit a correct passcode",
			input:  `{"code": "123456"}`,
			flowId: "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg: config.Config{
				Passcode: config.Passcode{
					Enabled: true,
				},
			},
			expectedState: shared.StateSuccess,
			statusCode:    http.StatusOK,
		},
		{
			name:          "submit a wrong passcode",
			input:         `{"code": "654321"}`,
			flowId:        "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg:           config.Config{},
			expectedState: StatePasscodeConfirmation,
			statusCode:    http.StatusUnauthorized,
		},
		{
			name:          "submit a passcode which max attempts are reached",
			input:         `{"code": "654321"}`,
			flowId:        "8a2cf90d-dea5-4678-9dca-6707dab6af77",
			cfg:           config.Config{},
			expectedState: StatePasscodeConfirmation,
			statusCode:    http.StatusUnauthorized,
		},
		{
			name:             "submit a passcode where the id is not in the stash",
			input:            `{"code": "123456"}`,
			flowId:           "23524801-f445-4859-bc16-22cf1dd417ac",
			cfg:              config.Config{},
			expectsFlowError: true,
			expectedState:    shared.StateError,
			statusCode:       http.StatusInternalServerError,
		},
		{
			name:             "submit a passcode where the passcode is not stored in the DB",
			input:            `{"code": "123456"}`,
			flowId:           "fc4dc7e4-bce7-4154-873b-cb3d766df279",
			expectsFlowError: true,
			cfg:              config.Config{},
			expectedState:    shared.StateError,
			statusCode:       http.StatusInternalServerError,
		},
		{
			name:          "submit a passcode where the passcode is expired",
			input:         `{"code": "123456"}`,
			flowId:        "5a862a2d-0d10-4904-b297-cb32fc43c859",
			cfg:           config.Config{},
			expectedState: StatePasscodeConfirmation,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:   "submit a correct passcode and passwords are enabled",
			input:  `{"code": "123456"}`,
			flowId: "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg: config.Config{
				Password: config.Password{
					Enabled: true,
				},
			},
			expectedState: shared.StatePasswordCreation,
			statusCode:    http.StatusOK,
		},
		{
			name:   "submit a correct passcode and passcodes for login are disabled and passkey onboarding is disabled",
			input:  `{"code": "123456"}`,
			flowId: "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg: config.Config{
				Passcode: config.Passcode{
					Enabled: false,
				},
				Passkey: config.Passkey{
					Onboarding: config.PasskeyOnboarding{
						Enabled: false,
					},
				},
			},
			expectedState: passkeyOnboarding.StateOnboardingCreatePasskey,
			statusCode:    http.StatusOK,
		},
		{
			name:   "submit a correct passcode and passkey onboarding is enabled and webauthn is available",
			input:  `{"code": "123456"}`,
			flowId: "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg: config.Config{
				Passcode: config.Passcode{
					Enabled: true,
				},
				Passkey: config.Passkey{
					Onboarding: config.PasskeyOnboarding{
						Enabled: true,
					},
				},
			},
			expectedState: passkeyOnboarding.StateOnboardingCreatePasskey,
			statusCode:    http.StatusOK,
		},
		{
			name:   "submit a correct passcode and passkey onboarding is enabled and webauthn is not available",
			input:  `{"code": "123456"}`,
			flowId: "bc3173e7-3204-4b9a-904b-9f812330b0de",
			cfg: config.Config{
				Passcode: config.Passcode{
					Enabled: true,
				},
				Passkey: config.Passkey{
					Onboarding: config.PasskeyOnboarding{
						Enabled: true,
					},
				},
			},
			expectedState: shared.StateSuccess,
			statusCode:    http.StatusOK,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../../test/fixtures/actions/submit_passcode")
			s.Require().NoError(err)

			passkeySubFlow, err := flowpilot.NewSubFlow().
				State(passkeyOnboarding.StateOnboardingCreatePasskey).
				Build()
			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(StatePasscodeConfirmation, VerifyPasscode{}).
				State(shared.StatePasswordCreation).
				State(shared.StateSuccess).
				State(shared.StateError).
				InitialState(StatePasscodeConfirmation).
				ErrorState(shared.StateError).
				SubFlows(passkeySubFlow).
				Build()
			s.Require().NoError(err)

			tx := s.Storage.GetConnection()
			db := models.NewFlowDB(tx)
			actionParam := fmt.Sprintf("submit_email_passcode@%s", currentTest.flowId)
			inputData := flowpilot.InputData{JSONString: currentTest.input}
			result, err := flow.Execute(db, flowpilot.WithActionParam(actionParam), flowpilot.WithInputData(inputData))
			if currentTest.expectsFlowError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Equal(currentTest.statusCode, result.Status())
				s.Equal(currentTest.expectedState, result.Response().StateName)
				// TODO: check that the schema of the action returns the correct error_code e.g.
				// result.Response().PublicActions[0].PublicSchema[0].PublicError.Code == ErrorValueInvalid
			}
		})
	}
}

func TestName(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), 12)
	log.Println(string(hash))
}
