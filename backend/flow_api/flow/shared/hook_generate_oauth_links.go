package shared

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"net/url"
)

type GenerateOAuthLinks struct {
	Action
}

func (h GenerateOAuthLinks) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	returnToUrl := deps.Cfg.ThirdParty.DefaultRedirectURL

	referer := deps.HttpContext.Request().Header.Get("Referer")
	if referer != "" {
		u, err := url.Parse(referer)
		if err != nil {
			return err
		}

		// remove any query and fragment parts of the referer
		u.RawQuery = ""
		u.Fragment = ""
		returnToUrl = u.String()
	}

	if deps.Cfg.ThirdParty.Providers.GitHub.Enabled {
		c.AddLink(OAuthLink("github", h.generateHref(deps.HttpContext, "github", returnToUrl)))
	}
	if deps.Cfg.ThirdParty.Providers.Google.Enabled {
		c.AddLink(OAuthLink("google", h.generateHref(deps.HttpContext, "google", returnToUrl)))
	}
	if deps.Cfg.ThirdParty.Providers.Apple.Enabled {
		c.AddLink(OAuthLink("apple", h.generateHref(deps.HttpContext, "apple", returnToUrl)))
	}

	return nil
}

func (h GenerateOAuthLinks) generateHref(c echo.Context, provider string, returnToUrl string) string {
	host := c.Request().Host
	forwardedProto := c.Request().Header.Get("X-Forwarded-Proto")
	if forwardedProto == "" {
		// Assume that a proxy is setting the X-Forwarded-Proto header correctly. Hanko should always be deployed behind a proxy,
		// because you cannot start the backend with https and passkeys only work in a secure context.
		// If the X-Forwarded-Proto header is not set, set it to 'http' because otherwise you would need to set up a https environment for local testing.
		forwardedProto = "http"
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s/thirdparty/auth", forwardedProto, host))
	query := url.Values{}
	query.Set("provider", provider)
	if returnToUrl != "" {
		query.Set("redirect_to", returnToUrl)
	}
	u.RawQuery = query.Encode()

	return u.String()
}
