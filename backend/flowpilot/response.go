package flowpilot

import (
	"fmt"
	"net/http"
)

// PublicAction represents a link to an action.
type PublicAction struct {
	Href         string       `json:"href"`
	PublicSchema PublicSchema `json:"inputs"`
	Name         ActionName   `json:"action"`
	Description  string       `json:"description"`
}

// PublicActions is a collection of PublicAction instances.
type PublicActions map[ActionName]PublicAction

// PublicError represents an error for public exposure.
type PublicError struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Cause   *string `json:"cause,omitempty"`
}

// PublicInput represents an input field for public exposure.
type PublicInput struct {
	Name          string        `json:"name"`
	Type          InputType     `json:"type"`
	Value         interface{}   `json:"value,omitempty"`
	MinLength     *int          `json:"min_length,omitempty"`
	MaxLength     *int          `json:"max_length,omitempty"`
	Required      *bool         `json:"required,omitempty"`
	Hidden        *bool         `json:"hidden,omitempty"`
	PublicError   *PublicError  `json:"error,omitempty"`
	AllowedValues AllowedValues `json:"allowed_values,omitempty"`
}

// PublicResponse represents the response of an action execution.
type PublicResponse struct {
	Name          StateName     `json:"name"`
	FlowPath      *string       `json:"flow_path,omitempty"`
	Status        int           `json:"status"`
	Payload       interface{}   `json:"payload,omitempty"`
	CSRFToken     string        `json:"csrf_token"`
	PublicActions PublicActions `json:"actions"`
	PublicError   *PublicError  `json:"error,omitempty"`
	PublicLinks   PublicLinks   `json:"links"`
}

// PublicLinks is a collection of Link instances.
type PublicLinks []PublicLink

// PublicLink represents a link for public exposure.
// A PublicLink can be a link to an oauth provider (e.g. google) or to the registration/login/tos/privacy page
type PublicLink struct {
	Name     string       `json:"name"` // tos, privacy, google, apple, microsoft, login, registration ... // how can we insert custom oauth provider here
	Href     string       `json:"href"`
	Category LinkCategory `json:"category"` // oauth, legal, other, ...
	Target   LinkTarget   `json:"target"`   // can be used to add the target of the a-tag e.g. _blank
}

// FlowResult interface defines methods for obtaining response and status.
type FlowResult interface {
	Response() PublicResponse
	Status() int
}

// DefaultFlowResult implements FlowResult interface.
type DefaultFlowResult struct {
	PublicResponse
}

// newFlowResultFromResponse creates a FlowResult from a PublicResponse.
func newFlowResultFromResponse(publicResponse PublicResponse) FlowResult {
	return DefaultFlowResult{PublicResponse: publicResponse}
}

// newFlowResultFromError creates a FlowResult from a FlowError.
func newFlowResultFromError(stateName StateName, flowError FlowError, debug bool) FlowResult {
	publicError := flowError.toPublicError(debug)
	status := flowError.Status()

	publicResponse := PublicResponse{
		Name:          stateName,
		Status:        status,
		PublicError:   &publicError,
		PublicActions: PublicActions{},
	}

	return DefaultFlowResult{PublicResponse: publicResponse}
}

// Response returns the PublicResponse.
func (r DefaultFlowResult) Response() PublicResponse {
	return r.PublicResponse
}

// Status returns the HTTP status code.
func (r DefaultFlowResult) Status() int {
	return r.PublicResponse.Status
}

// actionExecutionResult holds the result of a method execution.
type actionExecutionResult struct {
	actionName  ActionName
	schema      ExecutionSchema
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
func (er *executionResult) generateResponse(fc *defaultFlowContext, debug bool) FlowResult {
	// Generate actions for the response.
	actions := er.generateActions(fc)

	// Unmarshal the generated payload for the response.
	p := fc.payload.Unmarshal()

	// Generate links for the response.
	links := er.generateLinks()

	// Create the response object.
	resp := PublicResponse{
		Name:          er.nextStateName,
		Status:        http.StatusOK,
		Payload:       p,
		PublicActions: actions,
		PublicLinks:   links,
		CSRFToken:     fc.flowModel.CSRFToken,
	}

	if debug {
		fp := fc.GetFlowPath().String()
		resp.FlowPath = &fp
	}

	// Include flow error if present.
	if er.flowError != nil {
		status := er.flowError.Status()
		publicError := er.flowError.toPublicError(debug)

		resp.Status = status
		resp.PublicError = &publicError
	}

	return newFlowResultFromResponse(resp)
}

func (er *executionResult) generateLinks() PublicLinks {
	var publicLinks PublicLinks

	for _, link := range er.links {
		publicLink := link.toPublicLink()
		publicLinks = append(publicLinks, publicLink)
	}

	return publicLinks
}

// generateActions generates a collection of links based on the execution result.
func (er *executionResult) generateActions(fc *defaultFlowContext) PublicActions {
	var publicActions = make(PublicActions)

	// Get actions for the next addState.
	state, _ := fc.flow.getState(er.nextStateName)

	if state != nil {
		for _, actionDetail := range state.actionDetails {
			actionName := actionDetail.action.GetName()
			actionDescription := actionDetail.action.GetDescription()

			// Create action HREF based on the current flow context and method name.
			href := er.createHref(fc, actionName)
			schema := er.getSchema(fc, actionDetail)

			// (Re-)Initialize each action
			aic := defaultActionInitializationContext{
				schema:             schema.toInitializationSchema(),
				defaultFlowContext: fc,
			}

			actionDetail.action.Initialize(&aic)

			if aic.isSuspended {
				continue
			}

			publicSchema := schema.toPublicSchema(er.nextStateName)

			// Create the action instance.
			publicAction := PublicAction{
				Href:         href,
				PublicSchema: publicSchema,
				Name:         actionName,
				Description:  actionDescription,
			}

			publicActions[actionName] = publicAction
		}
	}

	return publicActions
}

// getSchema returns the schema for a given method name.
func (er *executionResult) getSchema(fc *defaultFlowContext, actionDetail defaultActionDetail) ExecutionSchema {
	if er.actionExecutionResult == nil ||
		actionDetail.action.GetName() != er.actionExecutionResult.actionName ||
		actionDetail.flowPath.String() != fc.GetFlowPath().String() || er.nextStateName != fc.GetCurrentState() {
		return newSchema()
	}
	return er.actionExecutionResult.schema
}

// createHref creates a link HREF based on the current flow context and method name.
func (er *executionResult) createHref(fc *defaultFlowContext, actionName ActionName) string {
	queryParam := createQueryParam(string(actionName), fc.GetFlowID())
	return fmt.Sprintf("%s?flowpilot_action=%s", fc.GetPath(), queryParam)
}
