package shared

import "github.com/teamhanko/hanko/backend/v3/flowpilot"

// Link categories enumeration.
const (
	CategoryLegal flowpilot.LinkCategory = "legal"
	CategoryOther flowpilot.LinkCategory = "other"
)

// LegalLink creates a new link with legal the category "legal".
func LegalLink(name string, href string) flowpilot.Link {
	return flowpilot.NewLink(name, CategoryLegal, href).Target(flowpilot.LinkTargetBlank)
}

// OtherLink creates a new link with legal the category "other".
func OtherLink(name string, href string) flowpilot.Link {
	return flowpilot.NewLink(name, CategoryOther, href)
}
