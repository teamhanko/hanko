package actions

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	passkeyOnboardingStates "github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"testing"
)

func TestSubmitNewPassword(t *testing.T) {
	s := new(submitNewPassword)

	suite.Run(t, s)
}

type submitNewPassword struct {
	test.Suite
}

func (s *submitNewPassword) TestSubmitNewPassword_Execute() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	tests := []struct {
		name               string
		input              string
		flowId             string
		cfg                config.Config
		expectedState      flowpilot.StateName
		expectedInputError flowpilot.InputError
		expectedFlowError  flowpilot.FlowError
		statusCode         int
	}{
		{
			name:               "submit a new password",
			input:              `{"new_password": "SuperSecure"}`,
			flowId:             "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg:                config.Config{},
			expectedState:      shared.StateSuccess,
			expectedInputError: nil,
			expectedFlowError:  nil,
			statusCode:         http.StatusOK,
		},
		{
			name:   "submit a new password that is too short",
			input:  `{"new_password": "test"}`,
			flowId: "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg: config.Config{
				Password: config.Password{
					MinPasswordLength: 8,
				},
			},
			expectedState:      shared.StatePasswordCreation,
			expectedInputError: flowpilot.ErrorValueTooShort,
			expectedFlowError:  flowpilot.ErrorFormDataInvalid,
			statusCode:         http.StatusBadRequest,
		},
		{
			name:               "submit a password that is too long",
			input:              `{"new_password": "ThisIsAVeryVeryLongPasswordToCheckTheLengthCheckAndItMustBeVeryLongInOrderToDoSo"}`,
			flowId:             "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg:                config.Config{},
			expectedState:      shared.StatePasswordCreation,
			expectedInputError: flowpilot.ErrorValueTooLong,
			expectedFlowError:  flowpilot.ErrorFormDataInvalid,
			statusCode:         http.StatusBadRequest,
		},
		{
			name:   "submit a new password and webauthn is not available and passkey onboarding is enabled",
			input:  `{"new_password": "SuperSecure"}`,
			flowId: "8a2cf90d-dea5-4678-9dca-6707dab6af77",
			cfg: config.Config{
				Passkey: config.Passkey{
					Onboarding: config.PasskeyOnboarding{
						Enabled: true,
					},
				},
			},
			expectedState:      shared.StateSuccess,
			expectedInputError: nil,
			expectedFlowError:  nil,
			statusCode:         http.StatusOK,
		},
		{
			name:   "submit a new password and webauthn is available and passkey onboarding is disabled",
			input:  `{"new_password": "SuperSecure"}`,
			flowId: "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg: config.Config{
				Passkey: config.Passkey{
					Onboarding: config.PasskeyOnboarding{
						Enabled: false,
					},
				},
			},
			expectedState:      shared.StateSuccess,
			expectedInputError: nil,
			expectedFlowError:  nil,
			statusCode:         http.StatusOK,
		},
		{
			name:   "submit a new password and webauthn is available and passkey onboarding is enabled",
			input:  `{"new_password": "SuperSecure"}`,
			flowId: "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg: config.Config{
				Passkey: config.Passkey{
					Onboarding: config.PasskeyOnboarding{
						Enabled: true,
					},
				},
			},
			expectedState:      passkeyOnboardingStates.StateOnboardingCreatePasskey,
			expectedInputError: nil,
			expectedFlowError:  nil,
			statusCode:         http.StatusOK,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../../test/fixtures/actions/submit_new_password")
			s.Require().NoError(err)

			passkeySubFlow, err := flowpilot.NewSubFlow().
				State(passkeyOnboardingStates.StateOnboardingCreatePasskey).
				Build()
			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(shared.StatePasswordCreation, NewSubmitNewPassword(currentTest.cfg)).
				State(shared.StateSuccess).
				SubFlows(passkeySubFlow).
				InitialState(shared.StatePasswordCreation).
				ErrorState(shared.StateError).
				Debug(true).
				Build()
			s.Require().NoError(err)

			tx := s.Storage.GetConnection()
			db := models.NewFlowDB(tx)
			actionParam := fmt.Sprintf("submit_new_password@%s", currentTest.flowId)
			inputData := flowpilot.InputData{JSONString: currentTest.input}
			result, err := flow.Execute(db, flowpilot.WithActionParam(actionParam), flowpilot.WithInputData(inputData))
			s.Require().NoError(err)

			s.Equal(currentTest.statusCode, result.Status())
			s.Equal(currentTest.expectedState, result.Response().StateName)

			if currentTest.expectedFlowError != nil {
				s.Equal(currentTest.expectedFlowError.Code(), result.Response().PublicError.Code)
			}

			if currentTest.expectedInputError != nil {
				s.Equal(currentTest.expectedInputError.Code(), result.Response().PublicActions[0].PublicSchema[0].PublicError.Code)
			}
		})
	}

}
