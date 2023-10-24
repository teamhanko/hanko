package actions

import (
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmitRegistrationIdentifier(t *testing.T) {
	s := new(submitRegistrationIdentifierActionSuite)

	suite.Run(t, s)
}

type submitRegistrationIdentifierActionSuite struct {
	test.Suite
}

func (s *submitRegistrationIdentifierActionSuite) TestSubmitRegistrationIdentifierExecute() {
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
			name:  "can register a new user with email",
			input: `{"email":"new@example.com"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled: "disabled",
					},
					Email: config.IdentifierEmail{
						Enabled: "required",
					},
				},
				Emails: config.Emails{
					RequireVerification: true,
				},
			},
			expectedState: common.StateEmailVerification,
			statusCode:    http.StatusOK,
		},
		{
			name:  "can register a new user with username",
			input: `{"username":"new_user"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "required",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "disabled",
					},
				},
			},
			expectedState: common.StateOnboardingCreatePasskey,
			statusCode:    http.StatusOK,
		},
		{
			name:  "can register a new user with email and username",
			input: `{"email":"new@exmaple.com","username":"new_user"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "required",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "required",
					},
				},
				Emails: config.Emails{
					RequireVerification: true,
				},
			},
			expectedState: common.StateEmailVerification,
			statusCode:    http.StatusOK,
		},
		{
			name:  "cannot register a new user with existing email",
			input: `{"email":"john.doe@example.com", "username": ""}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled: "optional",
					},
					Email: config.IdentifierEmail{
						Enabled: "required",
					},
				},
				Emails: config.Emails{
					RequireVerification: true,
				},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:  "do not return an error when user enumeration protection is implicit enabled",
			input: `{"email":"john.doe@example.com", "username": ""}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled: "disabled",
					},
					Email: config.IdentifierEmail{
						Enabled: "required",
					},
				},
				Emails: config.Emails{
					RequireVerification: true,
				},
			},
			expectedState: common.StateEmailVerification,
			statusCode:    http.StatusOK,
		},
		{
			name:  "cannot register a new user with existing username",
			input: `{"username":"john.doe"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "required",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "disabled",
					},
				},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:  "cannot register a new user missing required email",
			input: `{"username":"new_user"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "optional",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "required",
					},
				},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:  "cannot register a new user missing required username",
			input: `{"email":"new@example.com"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "required",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "optional",
					},
				},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:  "cannot register a new user with username with not allowed characters",
			input: `{"username": "unwanted ch@r@cters"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "required",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "disabled",
					},
				},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:  "cannot register a new user with too short username",
			input: `{"username": "t"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "required",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "disabled",
					},
				},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:  "cannot register a new user with too long username",
			input: `{"username":"this_is_a_very_very_long_username_to_check_if_this_username_is_rejected"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled:           "required",
						MaxLength:         64,
						MinLength:         3,
						AllowedCharacters: "abcdefghijklmnopqrstuvwxyz123456789-_.",
					},
					Email: config.IdentifierEmail{
						Enabled: "disabled",
					},
				},
			},
			expectedState: common.StateRegistrationInit,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:  "can register a new user with email verification disabled and password disabled",
			input: `{"email": "new@example.com"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled: "disabled",
					},
					Email: config.IdentifierEmail{
						Enabled: "required",
					},
				},
				Emails: config.Emails{
					RequireVerification: false,
				},
				Password: config.Password{
					Enabled: false,
				},
			},
			expectedState: common.StateOnboardingCreatePasskey,
			statusCode:    http.StatusOK,
		},
		{
			name:  "can register a new user with password enabled and email verification disabled",
			input: `{"email": "new@example.com"}`,
			cfg: config.Config{
				Identifier: config.Identifier{
					Username: config.IdentifierUsername{
						Enabled: "disabled",
					},
					Email: config.IdentifierEmail{
						Enabled: "required",
					},
				},
				Emails: config.Emails{
					RequireVerification: false,
				},
				Password: config.Password{
					Enabled: true,
				},
			},
			expectedState: common.StatePasswordCreation,
			statusCode:    http.StatusOK,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../../test/fixtures/actions/submit_registration_identifier")
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/passcode/login/initialize", nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			passkeySubFlow, err := flowpilot.NewSubFlow().
				State(common.StateOnboardingCreatePasskey).
				Build()

			s.Require().NoError(err)

			flow, err := flowpilot.NewFlow("/registration_test").
				State(common.StateRegistrationInit, NewSubmitRegistrationIdentifier(currentTest.cfg, s.Storage, &testPasscodeService{}, echo.New().NewContext(req, rec))).
				State(common.StateEmailVerification).
				State(common.StateSuccess).
				State(common.StatePasswordCreation).
				SubFlows(passkeySubFlow).
				InitialState(common.StateRegistrationInit).
				ErrorState(common.StateError).
				Build()

			s.Require().NoError(err)

			tx := s.Storage.GetConnection()
			db := models.NewFlowDB(tx)
			actionParam := "submit_registration_identifier@0b41f4dd-8e46-4a7c-bb4d-d60843113431"
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

// TODO: should be removed, tests should use new email test server instance introduced in https://github.com/teamhanko/hanko/pull/1093
type testPasscodeService struct {
}

func (t *testPasscodeService) SendEmailVerification(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error) {
	return uuid.NewV4()
}

func (t *testPasscodeService) SendLogin(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error) {
	return uuid.NewV4()
}

func (t *testPasscodeService) PasswordRecovery(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error) {
	return uuid.NewV4()
}
