package flowpilot

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
	"net/http"
)

// Link represents a link to an action.
type Link struct {
	Href        string       `json:"href"`
	Inputs      PublicSchema `json:"inputs"`
	MethodName  MethodName   `json:"method_name"`
	Description string       `json:"description"`
}

// Links is a collection of Link instances.
type Links []Link

// Add adds a link to the collection of Links.
func (ls *Links) Add(l Link) {
	*ls = append(*ls, l)
}

// PublicError represents an error for public exposure.
type PublicError struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Origin  *string `json:"origin,omitempty"`
}

// PublicInput represents an input field for public exposure.
type PublicInput struct {
	Name      string       `json:"name"`
	Type      InputType    `json:"type"`
	Value     interface{}  `json:"value,omitempty"`
	MinLength *int         `json:"min_length,omitempty"`
	MaxLength *int         `json:"max_length,omitempty"`
	Required  *bool        `json:"required,omitempty"`
	Hidden    *bool        `json:"hidden,omitempty"`
	Error     *PublicError `json:"error,omitempty"`
}

// PublicResponse represents the response of an action execution.
type PublicResponse struct {
	State   StateName    `json:"state"`
	Status  int          `json:"status"`
	Payload interface{}  `json:"payload,omitempty"`
	Links   Links        `json:"links"`
	Error   *PublicError `json:"error,omitempty"`
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
func newFlowResultFromResponse(response PublicResponse) FlowResult {
	return DefaultFlowResult{PublicResponse: response}
}

// newFlowResultFromError creates a FlowResult from a FlowError.
func newFlowResultFromError(stateName StateName, flowError FlowError, debug bool) FlowResult {
	pe := flowError.toPublicError(debug)

	return DefaultFlowResult{PublicResponse: PublicResponse{
		State:  stateName,
		Status: flowError.Status(),
		Error:  &pe,
	}}
}

// Response returns the PublicResponse.
func (r DefaultFlowResult) Response() PublicResponse {
	return r.PublicResponse
}

// Status returns the HTTP status code.
func (r DefaultFlowResult) Status() int {
	return r.PublicResponse.Status
}

// methodExecutionResult holds the result of a method execution.
type methodExecutionResult struct {
	methodName MethodName
	schema     MethodExecutionSchema
}

// executionResult holds the result of an action execution.
type executionResult struct {
	nextState StateName
	flowError FlowError

	*methodExecutionResult
}

// generateResponse generates a response based on the execution result.
func (er *executionResult) generateResponse(fc defaultFlowContext) FlowResult {
	// Generate links for the response.
	links := er.generateLinks(fc)

	// Create the response object.
	resp := PublicResponse{
		State:   er.nextState,
		Status:  http.StatusOK,
		Payload: fc.payload.Unmarshal(),
		Links:   links,
	}

	// Include flow error if present.
	if er.flowError != nil {
		status := er.flowError.Status()
		publicError := er.flowError.toPublicError(false)

		resp.Status = status
		resp.Error = &publicError
	}

	return newFlowResultFromResponse(resp)
}

// generateLinks generates a collection of links based on the execution result.
func (er *executionResult) generateLinks(fc defaultFlowContext) Links {
	var links Links

	// Get transitions for the next state.
	transitions := fc.flow.getTransitionsForState(er.nextState)

	if transitions != nil {
		for _, t := range *transitions {
			currentMethodName := t.Method.GetName()
			currentDescription := t.Method.GetDescription()

			// Create link HREF based on the current flow context and method name.
			href := er.createHref(fc, currentMethodName)
			schema := er.getExecutionSchema(currentMethodName)

			if schema == nil {
				// Create schema if not available.
				if schema = er.createSchema(fc, t.Method); schema == nil {
					continue
				}
			}

			// Create the link instance.
			link := Link{
				Href:        href,
				Inputs:      schema.toPublicSchema(er.nextState),
				MethodName:  currentMethodName,
				Description: currentDescription,
			}

			links.Add(link)
		}
	}

	return links
}

// createSchema creates an execution schema for a method if needed.
func (er *executionResult) createSchema(fc defaultFlowContext, method Method) MethodExecutionSchema {
	var schema MethodExecutionSchema
	var err error

	if er.methodExecutionResult != nil {
		data := er.methodExecutionResult.schema.getOutputData()
		schema, err = newSchemaWithOutputData(data)
	} else {
		schema = newSchema()
	}

	if err != nil {
		return nil
	}

	// Initialize the method.
	mic := defaultMethodInitializationContext{schema: schema.toInitializationSchema(), stash: fc.stash}
	method.Initialize(&mic)

	if mic.isSuspended {
		return nil
	}

	return schema
}

// getExecutionSchema gets the execution schema for a given method name.
func (er *executionResult) getExecutionSchema(methodName MethodName) MethodExecutionSchema {
	if er.methodExecutionResult == nil || methodName != er.methodExecutionResult.methodName {
		return nil
	}

	return er.methodExecutionResult.schema
}

// createHref creates a link HREF based on the current flow context and method name.
func (er *executionResult) createHref(fc defaultFlowContext, methodName MethodName) string {
	action := utils.CreateActionParam(string(methodName), fc.GetFlowID())
	return fmt.Sprintf("%s?flowpilot_action=%s", fc.GetPath(), action)
}
