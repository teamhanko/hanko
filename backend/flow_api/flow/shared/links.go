package shared

import "github.com/teamhanko/hanko/backend/flowpilot"

// Link categories enumeration.
const (
	CategoryLegal flowpilot.LinkCategory = "legal"
	CategoryOauth flowpilot.LinkCategory = "oauth"
	CategoryOther flowpilot.LinkCategory = "other"
)

// LegalLink creates a new link with legal the category "legal".
func LegalLink(name string, href string) flowpilot.Link {
	return flowpilot.NewLink(name, CategoryLegal, href).Target(flowpilot.LinkTargetBlank)
}

// OAuthLink creates a new link with legal the category "oauth".
func OAuthLink(name string, href string) flowpilot.Link {
	return flowpilot.NewLink(name, CategoryOauth, href)
}

// OtherLink creates a new link with legal the category "other".
func OtherLink(name string, href string) flowpilot.Link {
	return flowpilot.NewLink(name, CategoryOther, href)
}
