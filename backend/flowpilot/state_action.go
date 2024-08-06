package flowpilot

type actionDetail interface {
	getAction() Action
	getFlowName() FlowName
}

type defaultActionDetail struct {
	action   Action
	flowName FlowName
}

// actions represents a list of action
type defaultActionDetails []actionDetail

func (ad *defaultActionDetail) getAction() Action {
	return ad.action
}

func (ad *defaultActionDetail) getFlowName() FlowName {
	return ad.flowName
}
