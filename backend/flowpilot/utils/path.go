package utils

import (
	"strings"
)

type Path struct {
	fragments []string
}

func NewPath(path string) Path {
	return Path{fragments: strings.Split(path, ".")}
}

func (p *Path) Add(fragment string) {
	p.fragments = append(p.fragments, fragment)
}

func (p *Path) Remove() {
	if len(p.fragments) > 0 {
		p.fragments = p.fragments[:len(p.fragments)-1]
	}
}

func (p *Path) String() string {
	return strings.Join(p.fragments, ".")
}
