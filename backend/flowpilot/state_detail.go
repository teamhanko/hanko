package flowpilot

import "fmt"

type stateDetail interface {
	getName() StateName
	getFlow() stateActions
	getFlowName() FlowName
	getSubFlows() SubFlows
	getActionDetails() defaultActionDetails
	getActionDetail(actionName ActionName) (actionDetail, error)
}

// state represents details for a state, including the associated actions, available sub-flows and more.
type defaultStateDetail struct {
	name          StateName
	flowName      FlowName
	flow          stateActions
	subFlows      SubFlows
	actionDetails defaultActionDetails
}

func (sd *defaultStateDetail) getName() StateName {
	return sd.name
}

func (sd *defaultStateDetail) getFlow() stateActions {
	return sd.flow
}

func (sd *defaultStateDetail) getFlowName() FlowName {
	return sd.flowName
}

func (sd *defaultStateDetail) getSubFlows() SubFlows {
	return sd.subFlows
}

func (sd *defaultStateDetail) getActionDetails() defaultActionDetails {
	return sd.actionDetails
}

// getActionDetail returns the Action with the specified name.
func (sd *defaultStateDetail) getActionDetail(actionName ActionName) (actionDetail, error) {
	for _, ad := range sd.actionDetails {
		currentActionName := ad.getAction().GetName()

		if currentActionName == actionName {
			return ad, nil
		}
	}

	return nil, fmt.Errorf("action '%s' not found", actionName)
}

// stateDetails maps states to associated Actions, flows and sub-flows.
type stateDetails map[StateName]stateDetail
