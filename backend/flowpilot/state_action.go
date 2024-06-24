package flowpilot

type actionDetail interface {
	getAction() Action
	getFlowName() string
	getFlowPath() flowPath
}

type defaultActionDetail struct {
	action   Action
	flowName string
	flowPath flowPath
}

// actions represents a list of action
type defaultActionDetails []actionDetail

func (ad *defaultActionDetail) getAction() Action {
	return ad.action
}

func (ad *defaultActionDetail) getFlowName() string {
	return ad.flowName
}

func (ad *defaultActionDetail) getFlowPath() flowPath {
	return ad.flowPath
}
