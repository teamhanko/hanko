package utils

import (
	"golang.org/x/exp/slices"
	"strings"
)

type Path interface {
	Add(fragment string)
	Remove()
	String() string
	HasFragment(fragment string) bool
}

type defaultPath struct {
	fragments []string
}

func NewPath(path string) Path {
	return &defaultPath{fragments: strings.Split(path, ".")}
}

func (p *defaultPath) Add(fragment string) {
	p.fragments = append(p.fragments, fragment)
}

func (p *defaultPath) Remove() {
	if len(p.fragments) > 0 {
		p.fragments = p.fragments[:len(p.fragments)-1]
	}
}

func (p *defaultPath) String() string {
	return strings.Join(p.fragments, ".")
}

func (p *defaultPath) HasFragment(fragment string) bool {
	return slices.Contains(p.fragments, fragment)
}
