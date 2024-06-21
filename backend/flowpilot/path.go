package flowpilot

import (
	"golang.org/x/exp/slices"
	"strings"
)

type flowPath interface {
	String() string
	HasFragment(fragment string) bool

	add(fragment string)
	remove()
	copy() flowPath
}

type defaultFlowPath struct {
	fragments []string
}

func newFlowPath() flowPath {
	return &defaultFlowPath{fragments: make([]string, 0)}
}

func newFlowPathFromString(path string) flowPath {
	return &defaultFlowPath{fragments: strings.Split(path, ".")}
}

func (p *defaultFlowPath) add(fragment string) {
	p.fragments = append(p.fragments, fragment)
}

func (p *defaultFlowPath) remove() {
	if len(p.fragments) > 0 {
		p.fragments = p.fragments[:len(p.fragments)-1]
	}
}

func (p *defaultFlowPath) copy() flowPath {
	cp := make([]string, len(p.fragments))
	copy(cp, p.fragments)
	return &defaultFlowPath{fragments: cp}
}

func (p *defaultFlowPath) String() string {
	return strings.Join(p.fragments, ".")
}

func (p *defaultFlowPath) HasFragment(fragment string) bool {
	return slices.Contains(p.fragments, fragment)
}
