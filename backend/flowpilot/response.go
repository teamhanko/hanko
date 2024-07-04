package flowpilot

import (
	"fmt"
	"net/http"
)

// ResponseAction represents a link to an action.
type ResponseAction struct {
	Href        string         `json:"href"`
	Inputs      ResponseInputs `json:"inputs"`
	Name        ActionName     `json:"action"`
	Description string         `json:"description"`
}

// ResponseActions is a collection of ResponseAction instances.
type ResponseActions map[ActionName]ResponseAction

// ResponseError represents an error for public exposure.
type ResponseError struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Cause   *string `json:"cause,omitempty"`
}

type ResponseAllowedValue struct {
	Value interface{} `json:"value"`
	Text  string      `json:"name"`
}

type ResponseAllowedValues []*ResponseAllowedValue

// ResponseInput represents an input field for public exposure.
type ResponseInput struct {
	Name          string                 `json:"name"`
	Type          inputType              `json:"type"`
	Value         interface{}            `json:"value,omitempty"`
	MinLength     *int                   `json:"min_length,omitempty"`
	MaxLength     *int                   `json:"max_length,omitempty"`
	Required      *bool                  `json:"required,omitempty"`
	Hidden        *bool                  `json:"hidden,omitempty"`
	Error         *ResponseError         `json:"error,omitempty"`
	AllowedValues *ResponseAllowedValues `json:"allowed_values,omitempty"`
}

// ResponseLinks is a collection of Link instances.
type ResponseLinks []ResponseLink

// ResponseLink represents a link for public exposure.
type ResponseLink struct {
	Name     string       `json:"name"` // tos, privacy, google, apple, microsoft, login, registration ... // how can we insert custom oauth provider here
	Href     string       `json:"href"`
	Category LinkCategory `json:"category"` // oauth, legal, other, ...
	Target   LinkTarget   `json:"target"`   // can be used to add the target of the a-tag e.g. _blank
}

// Response represents the response of an action execution.
type Response struct {
	Name      StateName       `json:"name"`
	Status    int             `json:"status"`
	Payload   interface{}     `json:"payload,omitempty"`
	CSRFToken string          `json:"csrf_token"`
	Actions   ResponseActions `json:"actions"`
	Error     *ResponseError  `json:"error,omitempty"`
	Links     ResponseLinks   `json:"links"`
}

// FlowResult interface defines methods for obtaining response and status.
type FlowResult interface {
	GetResponse() Response
	GetStatus() int
}

// defaultFlowResult implements FlowResult interface.
type defaultFlowResult struct {
	response Response
}

// newFlowResultFromResponse creates a FlowResult from a Response.
func newFlowResultFromResponse(response Response) FlowResult {
	return defaultFlowResult{response: response}
}

// newFlowResultFromError creates a FlowResult from a FlowError.
func newFlowResultFromError(stateName StateName, flowError FlowError, debug bool) FlowResult {
	e := flowError.toResponseError(debug)
	status := flowError.Status()

	response := Response{
		Name:    stateName,
		Status:  status,
		Error:   e,
		Actions: ResponseActions{},
	}

	return defaultFlowResult{response: response}
}

// GetResponse returns the Response.
func (r defaultFlowResult) GetResponse() Response {
	return r.response
}

// GetStatus returns the HTTP status code.
func (r defaultFlowResult) GetStatus() int {
	return r.response.Status
}

// actionExecutionResult holds the result of a method execution.
type actionExecutionResult struct {
	actionName  ActionName
	inputSchema executionInputSchema
	isSuspended bool
}

// executionResult holds the result of an action execution.
type executionResult struct {
	nextStateName StateName
	flowError     FlowError
	links         []Link

	*actionExecutionResult
}

// generateResponse generates a response based on the execution result.
func (er *executionResult) generateResponse(fc *defaultFlowContext) FlowResult {
	// Generate actions for the response.
	actions := er.generateActions(fc)

	// Unmarshal the generated payload for the response.
	p := fc.payload.Unmarshal()

	// Generate links for the response.
	links := er.generateLinks()

	// Create the response object.
	resp := Response{
		Name:      er.nextStateName,
		Status:    http.StatusOK,
		Payload:   p,
		Actions:   actions,
		Links:     links,
		CSRFToken: fc.flowModel.CSRFToken,
	}

	// Include flow error if present.
	if er.flowError != nil {
		status := er.flowError.Status()
		e := er.flowError.toResponseError(fc.flow.debug)

		resp.Status = status
		resp.Error = e
	}

	return newFlowResultFromResponse(resp)
}

func (er *executionResult) generateLinks() ResponseLinks {
	var links ResponseLinks

	for _, link := range er.links {
		l := link.toResponseLink()
		links = append(links, l)
	}

	return links
}

// generateActions generates a collection of links based on the execution result.
func (er *executionResult) generateActions(fc *defaultFlowContext) ResponseActions {
	var actions = make(ResponseActions)

	// Get actions for the next addState.
	state, _ := fc.flow.getState(er.nextStateName)

	if state != nil {
		for _, ad := range state.getActionDetails() {
			actionName := ad.getAction().GetName()
			actionDescription := ad.getAction().GetDescription()

			// Create action HREF based on the current flow context and method name.
			href, _ := er.createHref(fc, actionName)
			inputSchema := er.getInputSchema(fc, ad)

			// (Re-)Initialize each action
			aic := defaultActionInitializationContext{
				inputSchema:        inputSchema.forInitializationContext(),
				defaultFlowContext: fc,
			}

			ad.getAction().Initialize(&aic)

			if aic.isSuspended {
				continue
			}

			inputSchemaResponse := inputSchema.toResponseInputs(er.nextStateName)

			// Create the action instance.
			action := ResponseAction{
				Href:        href,
				Inputs:      inputSchemaResponse,
				Name:        actionName,
				Description: actionDescription,
			}

			actions[actionName] = action
		}
	}

	return actions
}

// getInputSchema returns the inputSchema for a given method name.
func (er *executionResult) getInputSchema(fc *defaultFlowContext, actionDetail actionDetail) executionInputSchema {
	actionName := actionDetail.getAction().GetName()

	if er.actionExecutionResult == nil ||
		actionName != er.actionExecutionResult.actionName ||
		er.nextStateName != fc.flowModel.CurrentState {
		return newSchema()
	}
	return er.actionExecutionResult.inputSchema
}

// createHref creates a link HREF based on the current flow context and method name.
func (er *executionResult) createHref(fc *defaultFlowContext, actionName ActionName) (string, error) {
	q, err := newQueryParam(fc.flow.queryParamKey, createQueryParamValue(actionName, fc.GetFlowID()))
	return fmt.Sprintf("/%s?%s", fc.GetFlowName(), q.getURLValues().Encode()), err
}
