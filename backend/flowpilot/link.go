package flowpilot

// LinkCategory represents the category of the link.
type LinkCategory string

// LinkTarget represents the html target attribute.
type LinkTarget string

// Link targets enumeration.
const (
	LinkTargetSelf   LinkTarget = "_self"
	LinkTargetBlank  LinkTarget = "_blank"
	LinkTargetParent LinkTarget = "_parent"
	LinkTargetTop    LinkTarget = "_top"
)

// Link defines the interface for links.
type Link interface {
	Target(LinkTarget) Link

	toPublicLink() PublicLink
}

// defaultLink represents a link with its options.
type defaultLink struct {
	name     string
	href     string
	category LinkCategory
	target   LinkTarget
}

// Target sets the target attribute of the link.
func (l *defaultLink) Target(target LinkTarget) Link {
	l.target = target
	return l
}

func (l *defaultLink) toPublicLink() PublicLink {
	return PublicLink{
		Name:     l.name,
		Href:     l.href,
		Category: l.category,
		Target:   l.target,
	}
}

// NewLink creates a new defaultLink instance with provided parameters.
func NewLink(name string, category LinkCategory, href string) Link {
	return &defaultLink{
		name:     name,
		href:     href,
		category: category,
		target:   LinkTargetSelf,
	}
}
