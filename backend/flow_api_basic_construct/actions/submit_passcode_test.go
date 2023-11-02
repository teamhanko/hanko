package actions

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
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
			expectedState: common.StateSuccess,
			statusCode:    http.StatusOK,
		},
		{
			name:          "submit a wrong passcode",
			input:         `{"code": "654321"}`,
			flowId:        "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg:           config.Config{},
			expectedState: common.StateRegistrationPasscodeConfirmation,
			statusCode:    http.StatusUnauthorized,
		},
		{
			name:          "submit a passcode which max attempts are reached",
			input:         `{"code": "654321"}`,
			flowId:        "8a2cf90d-dea5-4678-9dca-6707dab6af77",
			cfg:           config.Config{},
			expectedState: common.StateRegistrationPasscodeConfirmation,
			statusCode:    http.StatusUnauthorized,
		},
		{
			name:             "submit a passcode where the id is not in the stash",
			input:            `{"code": "123456"}`,
			flowId:           "23524801-f445-4859-bc16-22cf1dd417ac",
			cfg:              config.Config{},
			expectsFlowError: true,
			expectedState:    common.StateError,
			statusCode:       http.StatusInternalServerError,
		},
		{
			name:             "submit a passcode where the passcode is not stored in the DB",
			input:            `{"code": "123456"}`,
			flowId:           "fc4dc7e4-bce7-4154-873b-cb3d766df279",
			expectsFlowError: true,
			cfg:              config.Config{},
			expectedState:    common.StateError,
			statusCode:       http.StatusInternalServerError,
		},
		{
			name:          "submit a passcode where the passcode is expired",
			input:         `{"code": "123456"}`,
			flowId:        "5a862a2d-0d10-4904-b297-cb32fc43c859",
			cfg:           config.Config{},
			expectedState: common.StateRegistrationPasscodeConfirmation,
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
			expectedState: common.StatePasswordCreation,
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
			expectedState: common.StateOnboardingCreatePasskey,
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
			expectedState: common.StateOnboardingCreatePasskey,
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
			expectedState: common.StateSuccess,
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
				State(common.StateOnboardingCreatePasskey).
				Build()
			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(common.StateRegistrationPasscodeConfirmation, NewSubmitPasscode(currentTest.cfg, s.Storage)).
				State(common.StatePasswordCreation).
				State(common.StateSuccess).
				State(common.StateError).
				InitialState(common.StateRegistrationPasscodeConfirmation).
				ErrorState(common.StateError).
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
