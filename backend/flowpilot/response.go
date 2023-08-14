package flowpilot

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/utils"
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

// Response represents the response of an action execution.
type Response struct {
	State   StateName   `json:"state"`
	Payload interface{} `json:"payload,omitempty"`
	Links   Links       `json:"links"`
	Error   *ErrorType  `json:"error,omitempty"`
}

// methodExecutionResult holds the result of a method execution.
type methodExecutionResult struct {
	methodName MethodName
	schema     MethodExecutionSchema
}

// executionResult holds the result of an action execution.
type executionResult struct {
	nextState StateName
	errType   *ErrorType

	*methodExecutionResult
}

// generateResponse generates a response based on the execution result.
func (er *executionResult) generateResponse(fc defaultFlowContext) (*Response, error) {
	// Generate links for the response.
	links := er.generateLinks(fc)

	// Create the response object.
	resp := &Response{
		State:   er.nextState,
		Payload: fc.payload.Unmarshal(),
		Links:   links,
		Error:   er.errType,
	}
	return resp, nil
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
