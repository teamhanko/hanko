package flow

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/flow_api/flow/profile"
	"github.com/teamhanko/hanko/backend/v3/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v3/flowpilot"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

const testQueryParamKey = "action"

// inMemoryFlowDB is a flowpilot.FlowDB that keeps the flows in memory, so that actions can run without a database.
type inMemoryFlowDB map[uuid.UUID]flowpilot.FlowModel

func (db inMemoryFlowDB) GetFlow(flowID uuid.UUID, _ uuid.UUID) (*flowpilot.FlowModel, error) {
	flowModel, ok := db[flowID]
	if !ok {
		return nil, errors.New("flow not found")
	}
	return &flowModel, nil
}

func (db inMemoryFlowDB) CreateFlow(flowModel flowpilot.FlowModel) error {
	db[flowModel.ID] = flowModel
	return nil
}

func (db inMemoryFlowDB) UpdateFlow(flowModel flowpilot.FlowModel) error {
	db[flowModel.ID] = flowModel
	return nil
}

// setSessionUser provides the session user that the actions of the profile flow expect.
type setSessionUser struct{}

func (setSessionUser) Execute(c flowpilot.HookExecutionContext) error {
	c.Set("session_user", &models.User{ID: uuid.Must(uuid.NewV4())})
	return nil
}

func TestThirdPartyRedirectValidation(t *testing.T) {
	actions := map[string]flowpilot.Action{
		"login and registration": shared.ThirdPartyOAuth{},
		"profile":                profile.ConnectThirdpartyOauthProvider{},
	}

	for name, action := range actions {
		t.Run(name, func(t *testing.T) {
			t.Run("rejects a redirect URL that is not allowed and reports it on the input", func(t *testing.T) {
				response := runThirdPartyAction(t, action, "https://not-allowed.example.com")

				assert.Equal(t, shared.StateLoginInit, response.Name)
				assert.Equal(t, http.StatusBadRequest, response.Status)
				require.NotNil(t, response.Error)
				assert.Equal(t, flowpilot.ErrorFormDataInvalid.Code(), response.Error.Code)

				input := response.Actions[action.GetName()].Inputs["redirect_to"]
				require.NotNil(t, input)
				require.NotNil(t, input.Error)
				assert.Equal(t, shared.ErrorRedirectURLNotAllowed.Code(), input.Error.Code)
			})

			t.Run("continues with a redirect URL that is allowed", func(t *testing.T) {
				response := runThirdPartyAction(t, action, "https://allowed.example.com/callback")

				assert.Equal(t, shared.StateThirdParty, response.Name)
				assert.Nil(t, response.Error)
			})
		})
	}
}

// runThirdPartyAction starts a flow that offers the given action and executes the action with the given redirect URL.
func runThirdPartyAction(t *testing.T, action flowpilot.Action, redirectTo string) flowpilot.Response {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.ThirdParty.RedirectURL = "https://hanko.example.com/thirdparty/callback"
	cfg.ThirdParty.ErrorRedirectURL = "https://allowed.example.com/error"
	cfg.ThirdParty.AllowedRedirectURLS = []string{"https://allowed.example.com/**"}
	cfg.ThirdParty.Providers.Google.Enabled = true
	cfg.ThirdParty.Providers.Google.ClientID = "client-id"
	cfg.ThirdParty.Providers.Google.Secret = "secret"
	require.NoError(t, cfg.PostProcess())

	testFlow, err := flowpilot.NewFlow("test").
		State(shared.StateLoginInit, action).
		State(shared.StateThirdParty).
		State(shared.StateError).
		InitialState(shared.StateLoginInit).
		ErrorState(shared.StateError).
		BeforeState(shared.StateLoginInit, setSessionUser{}).
		BeforeEachAction(setSessionUser{}).
		Build()
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	testFlow.Set("deps", &shared.Dependencies{
		Cfg:         *cfg,
		HttpContext: echo.New().NewContext(request, httptest.NewRecorder()),
	})

	db := inMemoryFlowDB{}

	result, err := testFlow.Execute(db, flowpilot.WithQueryParamKey(testQueryParamKey))
	require.NoError(t, err)
	initResponse := result.GetResponse()
	require.Contains(t, initResponse.Actions, action.GetName())

	href, err := url.Parse(initResponse.Actions[action.GetName()].Href)
	require.NoError(t, err)
	queryParamValue := href.Query().Get(testQueryParamKey)
	require.NotEmpty(t, queryParamValue)

	result, err = testFlow.Execute(db,
		flowpilot.WithQueryParamKey(testQueryParamKey),
		flowpilot.WithQueryParamValue(queryParamValue),
		flowpilot.WithInputData(flowpilot.InputData{
			InputDataMap: map[string]interface{}{"redirect_to": redirectTo, "provider": "google"},
			CSRFToken:    initResponse.CSRFToken,
		}),
	)
	require.NoError(t, err)

	return result.GetResponse()
}
