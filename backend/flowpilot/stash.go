package flowpilot

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

// Stash defines the interface for managing JSON data.
type Stash interface {
	jsonmanager.JSONManager

	getLastStateFromHistory() (state, unscheduledState *StateName, numOfScheduledStates *int64, err error)
	addStateToHistory(state StateName, unscheduledState *StateName, numOfScheduledStates *int64) error
	removeLastStateFromHistory() error
	addScheduledStates(scheduledStates ...StateName) error
	removeLastScheduledState() (*StateName, error)
}

// defaultStash implements the Stash interface.
type defaultStash struct {
	jsonmanager.JSONManager
}

// NewStash creates a new instance of Stash with empty JSON data.
func NewStash() Stash {
	return &defaultStash{JSONManager: jsonmanager.NewJSONManager()}
}

// NewStashFromString creates a new instance of Stash with the given JSON data.
func NewStashFromString(data string) (Stash, error) {
	jm, err := jsonmanager.NewJSONManagerFromString(data)
	return &defaultStash{JSONManager: jm}, err
}

// addStateToHistory adds a state to the history. Specify the values for unscheduledState and numOfScheduledStates to
// maintain the list of scheduled states if sub-flows are involved.
func (s *defaultStash) addStateToHistory(state StateName, unscheduledState *StateName, numOfScheduledStates *int64) error {
	// Create a new JSONManager to manage the history item
	historyItem := jsonmanager.NewJSONManager()

	// Get the last state from history
	lastState, _, _, err := s.getLastStateFromHistory()
	if err != nil {
		return err
	}

	// If the last state is the same as the new state, do not add it again
	if lastState != nil && *lastState == state {
		return nil
	}

	// Set the state in the history item
	if err = historyItem.Set("s", state); err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	// If numOfScheduledStates is provided and greater than 0, set it in the history item
	if numOfScheduledStates != nil && *numOfScheduledStates > 0 {
		if err = historyItem.Set("n", *numOfScheduledStates); err != nil {
			return fmt.Errorf("failed to set num_of_scheduled_states: %w", err)
		}
	}

	// If unscheduledState is provided, set it in the history item
	if unscheduledState != nil {
		if err = historyItem.Set("u", *unscheduledState); err != nil {
			return fmt.Errorf("failed to set unscheduled_state: %w", err)
		}
	}

	// Update the stashed history with the new history item
	if err = s.Set("_.state_history.-1", historyItem.Unmarshal()); err != nil {
		return fmt.Errorf("failed to update stashed history: %w", err)
	}

	return nil
}

// removeLastStateFromHistory removes the last state from history.
func (s *defaultStash) removeLastStateFromHistory() error {
	if err := s.Delete("_.state_history.-1"); err != nil {
		return err
	}

	return nil
}

// getLastStateFromHistory returns the last state, as well as the values for unscheduledState and numOfScheduledStates.
// These values indicate that further states have been added or removed from the list of scheduled states during the
// last state.
func (s *defaultStash) getLastStateFromHistory() (state, unscheduledState *StateName, numOfScheduledStates *int64, err error) {
	// Get the index of the last history item
	lastItemPosition := s.Get("_.state_history.#").Int() - 1

	// Retrieve the last history item
	lastHistoryItem := s.Get(fmt.Sprintf("_.state_history.%d", lastItemPosition))

	// If the last history item doesn't exist, return nil values and no error
	if !lastHistoryItem.Exists() {
		return nil, nil, nil, nil
	}

	// Check if the last history item is an object
	if !lastHistoryItem.IsObject() {
		return nil, nil, nil, errors.New("last history item is not an object")
	}

	// Check if the 's' field exists in the last history item
	if !lastHistoryItem.Get("s").Exists() {
		return nil, nil, nil, errors.New("last history item is missing a value for 'state'")
	}

	// Parse 's' field and assign it to the 'state' variable
	sn := StateName(lastHistoryItem.Get("s").String())
	state = &sn

	// Check if 'u' field exists in the last history item
	if lastHistoryItem.Get("u").Exists() {
		// Parse 'u' field and assign it to the 'unscheduledState' variable
		usn := StateName(lastHistoryItem.Get("u").String())
		unscheduledState = &usn
	}

	// Check if 'n' field exists in the last history item
	if lastHistoryItem.Get("n").Exists() {
		// Parse 'n' field and assign it to the 'numOfScheduledStates' variable
		n := lastHistoryItem.Get("n").Int()
		numOfScheduledStates = &n
	}

	// Return the parsed values
	return state, unscheduledState, numOfScheduledStates, nil
}

// addScheduledStates adds scheduled states.
func (s *defaultStash) addScheduledStates(scheduledStates ...StateName) error {
	// get the current sub-flow stack from the stash
	stack := s.Get("_.scheduled_states").Array()

	newStack := make([]StateName, len(stack))

	for index := range newStack {
		newStack[index] = StateName(stack[index].String())
	}

	// prepend the scheduledStates to the list of previously defined scheduled states
	newStack = append(scheduledStates, newStack...)

	if err := s.Set("_.scheduled_states", newStack); err != nil {
		return fmt.Errorf("failed to set scheduled_states: %w", err)
	}

	return nil
}

// removeLastScheduledState removes and returns the last scheduled state if present.
func (s *defaultStash) removeLastScheduledState() (*StateName, error) {
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
	nextState := newStack[0]
	newStack = newStack[1:]

	// stash the updated list of scheduled states
	if err := s.Set("_.scheduled_states", newStack); err != nil {
		return nil, fmt.Errorf("failed to stash scheduled states while ending the sub-flow: %w", err)
	}

	return &nextState, nil
}
