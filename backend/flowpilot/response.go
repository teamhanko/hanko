package flowpilot

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
	"net/http"
)

// PublicAction represents a link to an action.
type PublicAction struct {
	Href         string       `json:"href"`
	PublicSchema PublicSchema `json:"schema"`
	Name         ActionName   `json:"action"`
	Description  string       `json:"description"`
}

// PublicActions is a collection of PublicAction instances.
type PublicActions []PublicAction

// Add adds a link to the collection of PublicActions.
func (pa *PublicActions) Add(publicAction PublicAction) {
	*pa = append(*pa, publicAction)
}

// PublicError represents an error for public exposure.
type PublicError struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Cause   *string `json:"cause,omitempty"`
}

// PublicInput represents an input field for public exposure.
type PublicInput struct {
	Name        string       `json:"name"`
	Type        InputType    `json:"type"`
	Value       interface{}  `json:"value,omitempty"`
	MinLength   *int         `json:"min_length,omitempty"`
	MaxLength   *int         `json:"max_length,omitempty"`
	Required    *bool        `json:"required,omitempty"`
	Hidden      *bool        `json:"hidden,omitempty"`
	PublicError *PublicError `json:"error,omitempty"`
}

// PublicResponse represents the response of an action execution.
type PublicResponse struct {
	StateName     StateName     `json:"state"`
	Status        int           `json:"status"`
	Payload       interface{}   `json:"payload,omitempty"`
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
		StateName:   stateName,
		Status:      status,
		PublicError: &publicError,
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
func (er *executionResult) generateResponse(fc defaultFlowContext, debug bool) FlowResult {
	// Generate actions for the response.
	actions := er.generateActions(fc)

	// Unmarshal the generated payload for the response.
	payload := fc.payload.Unmarshal()

	// Generate links for the response.
	links := er.generateLinks()

	// Create the response object.
	resp := PublicResponse{
		StateName:     er.nextStateName,
		Status:        http.StatusOK,
		Payload:       payload,
		PublicActions: actions,
		PublicLinks:   links,
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
func (er *executionResult) generateActions(fc defaultFlowContext) PublicActions {
	var publicActions PublicActions

	// Get actions for the next addState.
	state, _ := fc.flow.getState(er.nextStateName)

	if state != nil {
		for _, action := range state.actions {
			actionName := action.GetName()
			actionDescription := action.GetDescription()

			// Create action HREF based on the current flow context and method name.
			href := er.createHref(fc, actionName)
			schema := er.getExecutionSchema(actionName)

			if schema == nil {
				// Create schema if not available.
				if schema = er.createSchema(fc, action); schema == nil {
					continue
				}
			} else if er.isSuspended {
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

			publicActions.Add(publicAction)
		}
	}

	return publicActions
}

// createSchema creates an execution schema for a method if needed.
func (er *executionResult) createSchema(fc defaultFlowContext, action Action) ExecutionSchema {
	var schema ExecutionSchema

	if er.actionExecutionResult != nil {
		var err error

		data := er.actionExecutionResult.schema.getOutputData()
		schema, err = newSchemaWithOutputData(data)

		if err != nil {
			return nil
		}
	} else {
		schema = newSchema()
	}

	// Initialize the action.
	aic := defaultActionInitializationContext{
		schema:        schema.toInitializationSchema(),
		stash:         fc.stash,
		contextValues: fc.flow.contextValues,
	}

	action.Initialize(&aic)

	if aic.isSuspended {
		return nil
	}

	return schema
}

// getExecutionSchema gets the execution schema for a given method name.
func (er *executionResult) getExecutionSchema(actionName ActionName) ExecutionSchema {
	if er.actionExecutionResult == nil || actionName != er.actionExecutionResult.actionName {
		return nil
	}

	return er.actionExecutionResult.schema
}

// createHref creates a link HREF based on the current flow context and method name.
func (er *executionResult) createHref(fc defaultFlowContext, actionName ActionName) string {
	action := utils.CreateActionParam(string(actionName), fc.GetFlowID())
	return fmt.Sprintf("%s?flowpilot_action=%s", fc.GetPath(), action)
}
