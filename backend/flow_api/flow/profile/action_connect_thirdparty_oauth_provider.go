package profile

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"

	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/thirdparty"
	"github.com/teamhanko/hanko/backend/v2/utils"
	"golang.org/x/oauth2"
)

type ConnectThirdpartyOauthProvider struct {
	shared.Action
}

func (a ConnectThirdpartyOauthProvider) GetName() flowpilot.ActionName {
	return shared.ActionConnectThirdpartyOauthProvider
}

func (a ConnectThirdpartyOauthProvider) GetDescription() string {
	return "Connect a third party provider via OAuth."
}

func (a ConnectThirdpartyOauthProvider) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	enabledThirdPartyProviders := deps.Cfg.ThirdParty.Providers.GetEnabled()
	enabledCustomThirdPartyProviders := deps.Cfg.ThirdParty.CustomProviders.GetEnabled()

	if len(enabledCustomThirdPartyProviders) == 0 && len(enabledThirdPartyProviders) == 0 {
		c.SuspendAction()
		return
	}

	providerInput := flowpilot.StringInput("provider").
		Hidden(true).
		Required(true)

	availableProvider := 0
	for _, provider := range enabledThirdPartyProviders {
		// Check if the user already has an identity with this provider
		// to avoid duplicates and only show providers that are not yet connected
		if !slices.ContainsFunc(userModel.Identities, func(identity models.Identity) bool {
			return identity.ProviderID == provider.ID
		}) {
			availableProvider += 1
			providerInput.AllowedValue(provider.DisplayName, provider.ID)
		}
	}
	slices.SortFunc(enabledCustomThirdPartyProviders, func(a, b config.CustomThirdPartyProvider) int {
		return cmp.Compare(a.DisplayName, b.DisplayName)
	})

	for _, provider := range enabledCustomThirdPartyProviders {
		// Check if the user already has an identity with this provider
		// to avoid duplicates and only show providers that are not yet connected
		if !slices.ContainsFunc(userModel.Identities, func(identity models.Identity) bool {
			return identity.ProviderID == provider.ID
		}) {
			availableProvider += 1
			providerInput.AllowedValue(provider.DisplayName, provider.ID)
		}
	}

	if availableProvider == 0 {
		c.SuspendAction()
		return
	}

	c.AddInputs(
		flowpilot.StringInput("redirect_to").Hidden(true).Required(true),
		providerInput,
		flowpilot.StringInput("code_verifier").Hidden(true),
	)
}

func (a ConnectThirdpartyOauthProvider) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	errorRedirectTo := deps.HttpContext.Request().Header.Get("Referer")
	if errorRedirectTo == "" {
		errorRedirectTo = deps.Cfg.ThirdParty.ErrorRedirectURL
	}

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	redirectTo := c.Input().Get("redirect_to").String()
	if ok := thirdparty.IsAllowedRedirect(deps.Cfg.ThirdParty, redirectTo); !ok {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	providerName := c.Input().Get("provider").String()

	provider, err := thirdparty.GetProvider(deps.Cfg.ThirdParty, providerName)
	if err != nil {
		return c.Error(flowpilot.ErrorFormDataInvalid.Wrap(err))
	}

	codeVerifier := c.Input().Get("code_verifier")
	state, err := thirdparty.GenerateState(
		&deps.Cfg,
		providerName,
		redirectTo,
		thirdparty.GenerateStateForFlowAPI(true),
		thirdparty.GenerateStateWithPKCECodeVerifier(codeVerifier.String()),
		thirdparty.GenerateStateForLoggedInUser(userModel.ID),
	)
	if err != nil {
		return c.Error(flowpilot.ErrorTechnical.Wrap(err))
	}

	var opts []oauth2.AuthCodeOption
	if codeVerifier.Exists() {
		opts = append(opts, oauth2.S256ChallengeOption(codeVerifier.String()))
	}
	authCodeUrl := provider.AuthCodeURL(string(state), opts...)

	cookie := &http.Cookie{
		Name:     utils.HankoThirdpartyStateCookie,
		Value:    string(state),
		Path:     "/",
		Domain:   deps.Cfg.Session.Cookie.Domain,
		MaxAge:   300,
		Secure:   true,
		HttpOnly: deps.Cfg.Session.Cookie.HttpOnly,
		SameSite: http.SameSiteNoneMode,
	}

	deps.HttpContext.SetCookie(cookie)

	if err = c.Payload().Set("redirect_url", authCodeUrl); err != nil {
		return fmt.Errorf("failed to set redirect_url to payload: %w", err)
	}

	return c.Continue(shared.StateThirdParty)
}
