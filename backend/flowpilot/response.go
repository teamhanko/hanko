package flowpilot

import (
	"fmt"
	"hanko_flowsc/flowpilot/jsonmanager"
	"hanko_flowsc/flowpilot/utils"
)

// Link represents a link to an action.
type Link struct {
	Href        string       `json:"href"`
	Inputs      PublicInputs `json:"inputs"`
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
	schema     ResponseSchema
	inputData  jsonmanager.ReadOnlyJSONManager
}

// executionResult holds the result of an action execution.
type executionResult struct {
	nextState StateName
	errType   *ErrorType

	*methodExecutionResult
}

// generateResponse generates a response based on the execution result.
func (er *executionResult) generateResponse(fc defaultFlowContext) (*Response, error) {
	var links Links

	transitions := fc.flow.getTransitionsForState(er.nextState)

	if transitions != nil {
		for _, t := range *transitions {
			currentMethodName := t.Method.GetName()
			currentDescription := t.Method.GetDescription()

			href := er.createHref(fc, currentMethodName)

			var schema ResponseSchema

			if schema = er.getExecutionSchema(currentMethodName); schema == nil {
				defaultSchema := DefaultSchema{}
				mic := defaultMethodInitializationContext{schema: &defaultSchema, stash: fc.stash}

				t.Method.Initialize(&mic)

				if mic.isSuspended {
					continue
				}

				schema = &defaultSchema
			}

			schema.applyFlash(fc.flash)
			schema.applyStash(fc.stash)

			link := Link{
				Href:        href,
				Inputs:      schema.toPublicInputs(),
				MethodName:  currentMethodName,
				Description: currentDescription,
			}

			links.Add(link)
		}
	}

	resp := &Response{
		State:   er.nextState,
		Payload: fc.payload.Unmarshal(),
		Links:   links,
		Error:   er.errType,
	}

	return resp, nil
}

// getExecutionSchema gets the execution schema for a given method name.
func (er *executionResult) getExecutionSchema(methodName MethodName) ResponseSchema {
	if er.methodExecutionResult == nil || methodName != er.methodExecutionResult.methodName {
		// The current method result does not belong to the methodName.
		return nil
	}

	schema := er.methodExecutionResult.schema
	inputData := er.methodExecutionResult.inputData

	schema.preserveInputData(inputData)

	return schema
}

// createHref creates a link HREF based on the current flow context and method name.
func (er *executionResult) createHref(fc defaultFlowContext, methodName MethodName) string {
	action := utils.CreateActionParam(string(methodName), fc.GetFlowID())
	return fmt.Sprintf("%s?flowpilot_action=%s", fc.GetPath(), action)
}
