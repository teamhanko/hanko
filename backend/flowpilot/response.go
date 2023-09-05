package flowpilot

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
	"net/http"
)

// PublicAction represents a link to an action.
type PublicAction struct {
	Href        string       `json:"href"`
	Inputs      PublicSchema `json:"inputs"`
	ActionName  ActionName   `json:"action_name"`
	Description string       `json:"description"`
}

// PublicActions is a collection of PublicAction instances.
type PublicActions []PublicAction

// Add adds a link to the collection of PublicActions.
func (ls *PublicActions) Add(l PublicAction) {
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
	State   StateName     `json:"state"`
	Status  int           `json:"status"`
	Payload interface{}   `json:"payload,omitempty"`
	Actions PublicActions `json:"actions"`
	Error   *PublicError  `json:"error,omitempty"`
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

// actionExecutionResult holds the result of a method execution.
type actionExecutionResult struct {
	actionName ActionName
	schema     ExecutionSchema
}

// executionResult holds the result of an action execution.
type executionResult struct {
	nextState StateName
	flowError FlowError

	*actionExecutionResult
}

// generateResponse generates a response based on the execution result.
func (er *executionResult) generateResponse(fc defaultFlowContext) FlowResult {
	// Generate actions for the response.
	actions := er.generateActions(fc)

	// Create the response object.
	resp := PublicResponse{
		State:   er.nextState,
		Status:  http.StatusOK,
		Payload: fc.payload.Unmarshal(),
		Actions: actions,
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

// generateActions generates a collection of links based on the execution result.
func (er *executionResult) generateActions(fc defaultFlowContext) PublicActions {
	var actions PublicActions

	// Get transitions for the next state.
	transitions := fc.flow.getTransitionsForState(er.nextState)

	if transitions != nil {
		for _, t := range *transitions {
			currentActionName := t.Action.GetName()
			currentDescription := t.Action.GetDescription()

			// Create action HREF based on the current flow context and method name.
			href := er.createHref(fc, currentActionName)
			schema := er.getExecutionSchema(currentActionName)

			if schema == nil {
				// Create schema if not available.
				if schema = er.createSchema(fc, t.Action); schema == nil {
					continue
				}
			}

			// Create the action instance.
			action := PublicAction{
				Href:        href,
				Inputs:      schema.toPublicSchema(er.nextState),
				ActionName:  currentActionName,
				Description: currentDescription,
			}

			actions.Add(action)
		}
	}

	return actions
}

// createSchema creates an execution schema for a method if needed.
func (er *executionResult) createSchema(fc defaultFlowContext, method Action) ExecutionSchema {
	var schema ExecutionSchema
	var err error

	if er.actionExecutionResult != nil {
		data := er.actionExecutionResult.schema.getOutputData()
		schema, err = newSchemaWithOutputData(data)
	} else {
		schema = newSchema()
	}

	if err != nil {
		return nil
	}

	// Initialize the method.
	mic := defaultActionInitializationContext{schema: schema.toInitializationSchema(), stash: fc.stash}
	method.Initialize(&mic)

	if mic.isSuspended {
		return nil
	}

	return schema
}

// getExecutionSchema gets the execution schema for a given method name.
func (er *executionResult) getExecutionSchema(methodName ActionName) ExecutionSchema {
	if er.actionExecutionResult == nil || methodName != er.actionExecutionResult.actionName {
		return nil
	}

	return er.actionExecutionResult.schema
}

// createHref creates a link HREF based on the current flow context and method name.
func (er *executionResult) createHref(fc defaultFlowContext, methodName ActionName) string {
	action := utils.CreateActionParam(string(methodName), fc.GetFlowID())
	return fmt.Sprintf("%s?flowpilot_action=%s", fc.GetPath(), action)
}
