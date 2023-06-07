package thirdparty

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"net/url"
	"strings"
)

func IsAllowedRedirect(config config.ThirdParty, redirectTo string) bool {
	if redirectTo == "" {
		return false
	}

	redirectTo = strings.TrimSuffix(redirectTo, "/")

	for _, pattern := range config.AllowedRedirectURLMap {
		if pattern.Match(redirectTo) {
			return true
		}
	}

	return false
}

func GetErrorUrl(redirectTo string, err error) string {
	var redirectUrl string
	switch v := err.(type) {
	case *ThirdPartyError:
		redirectUrl = fmt.Sprintf("%s?%s", redirectTo, v.Query())
	default:
		u := url.Values{}
		u.Add("error", ErrorCodeServerError)
		u.Add("error_description", "an internal error has occurred")
		redirectUrl = fmt.Sprintf("%s?%s", redirectTo, u.Encode())
	}
	return redirectUrl
}
