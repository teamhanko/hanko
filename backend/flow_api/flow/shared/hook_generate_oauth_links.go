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
		// If X-Forwarded-Proto is not set, set it to 'http' and hope that an HTTPS redirect is in place.
		// It cannot be set to 'https' because local testing would be more complicated as you would need to set up https.
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
