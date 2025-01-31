package shared

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"github.com/teamhanko/hanko/backend/utils"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
	"net/http"
)

type ThirdPartyOAuth struct {
	Action
}

func (a ThirdPartyOAuth) GetName() flowpilot.ActionName {
	return ActionThirdPartyOAuth
}

func (a ThirdPartyOAuth) GetDescription() string {
	return "Sign up/sign in with a third party provider via OAuth."
}

func (a ThirdPartyOAuth) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	enabledThirdPartyProviders := deps.Cfg.ThirdParty.Providers.GetEnabled()
	enabledCustomThirdPartyProviders := deps.Cfg.ThirdParty.CustomProviders.GetEnabled()

	if len(enabledCustomThirdPartyProviders) == 0 && len(enabledThirdPartyProviders) == 0 {
		c.SuspendAction()
		return
	}

	providerInput := flowpilot.StringInput("provider").
		Hidden(true).
		Required(true)

	for _, provider := range enabledThirdPartyProviders {
		providerInput.AllowedValue(provider.DisplayName, provider.ID)
	}

	slices.SortFunc(enabledCustomThirdPartyProviders, func(a, b config.CustomThirdPartyProvider) bool {
		return a.DisplayName < b.DisplayName
	})

	for _, provider := range enabledCustomThirdPartyProviders {
		providerInput.AllowedValue(provider.DisplayName, provider.ID)
	}

	c.AddInputs(flowpilot.StringInput("redirect_to").Hidden(true).Required(true), providerInput)
}

func (a ThirdPartyOAuth) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

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

	state, err := thirdparty.GenerateState(&deps.Cfg, providerName, redirectTo, thirdparty.GenerateStateForFlowAPI(true))
	if err != nil {
		return c.Error(flowpilot.ErrorTechnical.Wrap(err))
	}

	authCodeUrl := provider.AuthCodeURL(string(state), oauth2.SetAuthURLParam("prompt", "consent"))

	//cookie := utils.GenerateStateCookie(&deps.Cfg, utils.HankoThirdpartyStateCookie, string(state), utils.CookieOptions{
	//	MaxAge:   300,
	//	Path:     "/",
	//	SameSite: http.SameSiteLaxMode,
	//})

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

	return c.Continue(StateThirdParty)
}
