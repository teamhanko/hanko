package flowpilot

import (
	"strings"
)

type flowPath interface {
	String() string

	add(FlowName)
	remove()
	copy() flowPath
}

type defaultFlowPath struct {
	flowNames []FlowName
}

func newFlowPath() flowPath {
	return &defaultFlowPath{flowNames: make([]FlowName, 0)}
}

func newFlowPathFromName(name FlowName) flowPath {
	return &defaultFlowPath{flowNames: []FlowName{name}}
}

func newFlowPathFromPath(path string) flowPath {
	parts := strings.Split(path, ".")
	names := make([]FlowName, len(parts))
	for i, _ := range names {
		names[i] = FlowName(parts[i])
	}
	return &defaultFlowPath{flowNames: names}
}

func (p *defaultFlowPath) add(name FlowName) {
	p.flowNames = append(p.flowNames, name)
}

func (p *defaultFlowPath) remove() {
	if len(p.flowNames) > 0 {
		p.flowNames = p.flowNames[:len(p.flowNames)-1]
	}
}

func (p *defaultFlowPath) copy() flowPath {
	cp := make([]FlowName, len(p.flowNames))
	copy(cp, p.flowNames)
	return &defaultFlowPath{flowNames: cp}
}

func (p *defaultFlowPath) String() string {
	parts := make([]string, len(p.flowNames))
	for i, _ := range parts {
		parts[i] = string(p.flowNames[i])
	}
	return strings.Join(parts, ".")
}
