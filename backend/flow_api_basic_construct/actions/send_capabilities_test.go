package actions

import (
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"testing"
)

func TestSendCapabilitiesAction(t *testing.T) {
	s := new(sendCapabilitiesActionSuite)

	suite.Run(t, s)
}

type sendCapabilitiesActionSuite struct {
	test.Suite
}

func (s *sendCapabilitiesActionSuite) TestSendCapabilitiesExecute() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	tests := []struct {
		name          string
		input         string
		cfg           config.Config
		expectedState flowpilot.StateName
		statusCode    int
	}{
		{
			name:  "webauthn available, passcode disabled, password disabled",
			input: `{"capabilities":"{\"webauthn\": {\"available\": true}}"}`,
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: "disabled"},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusOK,
		},
		{
			name:  "webauthn not available, passcode disabled, password disabled",
			input: `{"capabilities":"{\"webauthn\": {\"available\": false}}"}`,
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: "disabled"},
			},
			expectedState: common.StateError,
			statusCode:    http.StatusOK,
		},
		{
			name:  "webauthn not available, 2FA required & only security_key is allowed",
			input: `{"capabilities":"{\"webauthn\": {\"available\": false}}"}`,
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: "disabled"},
			},
			expectedState: common.StateError,
			statusCode:    http.StatusOK,
		},
		{
			name:  "no input data",
			input: "",
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: "disabled"},
			},
			expectedState: common.StateRegistrationPreflight,
			statusCode:    http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../../test/fixtures/actions/send_capabilities")
			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(common.StateRegistrationPreflight, NewSendCapabilities(currentTest.cfg)).
				State(common.StateRegistrationInit).
				State(common.StateError).
				State(common.StateSuccess).
				InitialState(common.StateRegistrationPreflight).
				ErrorState(common.StateError).
				Build()
			s.Require().NoError(err)

			tx := s.Storage.GetConnection()
			db := models.NewFlowDB(tx)
			actionParam := "send_capabilities@0b41f4dd-8e46-4a7c-bb4d-d60843113431"
			inputData := flowpilot.InputData{JSONString: currentTest.input}
			result, err := flow.Execute(db, flowpilot.WithActionParam(actionParam), flowpilot.WithInputData(inputData))

			if s.NoError(err) {
				s.Equal(currentTest.statusCode, result.Status())
				s.Equal(currentTest.expectedState, result.Response().StateName)
			}
		})
	}
}
