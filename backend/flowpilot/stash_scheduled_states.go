package flowpilot

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

type scheduledStatesStash interface {
	addScheduledStates(scheduledStateNames ...StateName) error
	removeLastScheduledState() (*StateName, error)
	getLastScheduledState() *StateName
}

type defaultScheduledStatesStash struct {
	jsonmanager.JSONManager
}

// addScheduledStates adds scheduled states.
func (s *defaultScheduledStatesStash) addScheduledStates(scheduledStateNames ...StateName) error {
	// get the current sub-flow stack from the stash
	stack := s.Get("_.scheduled_states").Array()

	newStack := make([]StateName, len(stack))

	for index := range newStack {
		newStack[index] = StateName(stack[index].String())
	}

	// prepend the scheduledStates to the list of previously defined scheduled states
	newStack = append(scheduledStateNames, newStack...)

	if err := s.Set("_.scheduled_states", newStack); err != nil {
		return fmt.Errorf("failed to set scheduled_states: %w", err)
	}

	return nil
}

func (s *defaultScheduledStatesStash) getLastScheduledState() *StateName {
	// retrieve the previously scheduled states form the stash
	stack := s.Get("_.scheduled_states").Array()
	if len(stack) == 0 {
		return nil
	}
	nextStateName := StateName(stack[0].String())
	return &nextStateName
}

// removeLastScheduledState removes and returns the last scheduled stateDetail if present.
func (s *defaultScheduledStatesStash) removeLastScheduledState() (*StateName, error) {
	// retrieve the previously scheduled states form the stash
	stack := s.Get("_.scheduled_states").Array()

	newStack := make([]StateName, len(stack))

	for index := range newStack {
		newStack[index] = StateName(stack[index].String())
	}

	if len(newStack) == 0 {
		return nil, nil
	}

	// get and remove first stack item
	nextStateName := newStack[0]
	newStack = newStack[1:]

	// stash the updated list of scheduled states
	if err := s.Set("_.scheduled_states", newStack); err != nil {
		return nil, fmt.Errorf("failed to stash scheduled states while ending the sub-flow: %w", err)
	}

	return &nextStateName, nil
}
