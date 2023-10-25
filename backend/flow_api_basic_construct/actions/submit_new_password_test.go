package actions

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"log"
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
		name          string
		input         string
		flowId        string
		cfg           config.Config
		expectedState flowpilot.StateName
		statusCode    int
	}{
		{
			name:          "submit a new password",
			input:         `{"new_password": "SuperSecure"}`,
			flowId:        "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg:           config.Config{},
			expectedState: common.StateSuccess,
			statusCode:    http.StatusOK,
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
			expectedState: common.StatePasswordCreation,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:          "submit a password that is too long",
			input:         `{"new_password": "ThisIsAVeryVeryLongPasswordToCheckTheLengthCheckAndItMustBeVeryLongInOrderToDoSo"}`,
			flowId:        "0b41f4dd-8e46-4a7c-bb4d-d60843113431",
			cfg:           config.Config{},
			expectedState: common.StatePasswordCreation,
			statusCode:    http.StatusBadRequest,
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
			expectedState: common.StateSuccess,
			statusCode:    http.StatusOK,
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
			expectedState: common.StateSuccess,
			statusCode:    http.StatusOK,
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
			expectedState: common.StateOnboardingCreatePasskey,
			statusCode:    http.StatusOK,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../../test/fixtures/actions/submit_new_password")
			s.Require().NoError(err)

			passkeySubFlow, err := flowpilot.NewSubFlow().
				State(common.StateOnboardingCreatePasskey).
				Build()
			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(common.StatePasswordCreation, NewSubmitNewPassword(currentTest.cfg)).
				State(common.StateSuccess).
				SubFlows(passkeySubFlow).
				InitialState(common.StatePasswordCreation).
				ErrorState(common.StateError).
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

			log.Println(result.Response().PublicError)
			// TODO: check that the schema of the action returns the correct error_code e.g.
			// result.Response().PublicActions[0].PublicSchema[0].PublicError.Code == ErrorValueInvalid
		})
	}

}
