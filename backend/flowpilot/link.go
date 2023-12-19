package flowpilot

// LinkCategory represents the category of the link.
type LinkCategory string

// Link categories enumeration.
const (
	CategoryLegal LinkCategory = "legal"
	CategoryOauth LinkCategory = "oauth"
	CategoryOther LinkCategory = "other"
)

// LinkTarget represents the html target attribute.
type LinkTarget string

// Link targets enumeration.
const (
	TargetSelf   = "_self"
	TargetBlank  = "_blank"
	TargetParent = "_parent"
	TargetTop    = "_top"
)

// Link defines the interface for links.
type Link interface {
	Target(LinkTarget) Link
	Href(string) Link

	toPublicLink() PublicLink
}

// DefaultLink represents a link with its options.
type DefaultLink struct {
	name     string
	href     string
	category LinkCategory
	target   LinkTarget
}

// Target sets the target attribute of the link.
func (l *DefaultLink) Target(target LinkTarget) Link {
	l.target = target
	return l
}

// Href sets the href attribute of the link.
func (l *DefaultLink) Href(href string) Link {
	l.href = href
	return nil
}

func (l *DefaultLink) toPublicLink() PublicLink {
	return PublicLink{
		Name:     l.name,
		Href:     l.href,
		Category: l.category,
		Target:   l.target,
	}
}

// newLink creates a new DefaultLink instance with provided parameters.
func newLink(name string, category LinkCategory, href string, target LinkTarget) Link {
	return &DefaultLink{
		name:     name,
		href:     href,
		category: category,
		target:   target,
	}
}

// LegalLink creates a new link with legal the category "legal".
func LegalLink(name string, href string) Link {
	return newLink(name, CategoryLegal, href, TargetBlank)
}

// OAuthLink creates a new link with legal the category "oauth".
func OAuthLink(name string, href string) Link {
	return newLink(name, CategoryOauth, href, TargetSelf)
}

// OtherLink creates a new link with legal the category "other".
func OtherLink(name string, href string) Link {
	return newLink(name, CategoryOther, href, TargetSelf)
}
