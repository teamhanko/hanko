package capabilities

import (
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
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

func (s *sendCapabilitiesActionSuite) TestSendCapabilities_Execute() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	stateRegistrationInit := flowpilot.StateName("registration_init")

	tests := []struct {
		name          string
		input         string
		cfg           config.Config
		expectedState flowpilot.StateName
		statusCode    int
	}{
		{
			name:  "webauthn available, passcode disabled, password disabled",
			input: `{"webauthn_available": "true"}`,
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: false},
			},
			expectedState: stateRegistrationInit,
			statusCode:    http.StatusOK,
		},
		{
			name:  "webauthn not available, passcode disabled, password disabled",
			input: `{"webauthn_available": "false"}`,
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: false},
			},
			expectedState: shared.StateError,
			statusCode:    http.StatusOK,
		},
		{
			name:  "webauthn not available, 2FA required & only security_key is allowed",
			input: `{"webauthn_available": "false"}`,
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: false},
			},
			expectedState: shared.StateError,
			statusCode:    http.StatusOK,
		},
		{
			name:  "no input data",
			input: "",
			cfg: config.Config{
				Password:     config.Password{Enabled: false},
				Passcode:     config.Passcode{Enabled: false},
				SecondFactor: config.SecondFactor{Enabled: false},
			},
			expectedState: StatePreflight,
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
				State(StatePreflight, RegisterClientCapabilities{}).
				State(stateRegistrationInit).
				State(shared.StateError).
				State(shared.StateSuccess).
				InitialState(StatePreflight).
				ErrorState(shared.StateError).
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
