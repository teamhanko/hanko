package shared

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"github.com/teamhanko/hanko/backend/utils"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
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

	enabledProviders := deps.Cfg.ThirdParty.Providers.GetEnabled()
	if len(enabledProviders) == 0 {
		c.SuspendAction()
		return
	}

	providerInput := flowpilot.StringInput("provider").
		Hidden(true).
		Required(true)

	for _, provider := range enabledProviders {
		providerInput.AllowedValue(strings.ToLower(provider.DisplayName), provider.DisplayName)
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
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	redirectTo := c.Input().Get("redirect_to").String()
	if ok := thirdparty.IsAllowedRedirect(deps.Cfg.ThirdParty, redirectTo); !ok {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	provider, err := thirdparty.GetProvider(deps.Cfg.ThirdParty, c.Input().Get("provider").String())
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid.Wrap(err))
	}

	state, err := thirdparty.GenerateState(&deps.Cfg, provider.Name(), redirectTo, thirdparty.GenerateStateForFlowAPI(true))
	if err != nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorTechnical.Wrap(err))
	}

	authCodeUrl := provider.AuthCodeURL(string(state), oauth2.SetAuthURLParam("prompt", "consent"))

	cookie := utils.GenerateStateCookie(&deps.Cfg, utils.HankoThirdpartyStateCookie, string(state), utils.CookieOptions{
		MaxAge:   300,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	deps.HttpContext.SetCookie(cookie)

	if err = c.Payload().Set("redirect_url", authCodeUrl); err != nil {
		return fmt.Errorf("failed to set redirect_url to payload: %w", err)
	}

	return c.ContinueFlow(StateThirdPartyOAuth)
}

func (a ThirdPartyOAuth) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
