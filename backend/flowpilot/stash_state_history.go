package flowpilot

import (
	"errors"
	"fmt"
)

type stateHistoryStash interface {
	getLastStateFromHistory() (stateName, unscheduledState *StateName, numOfScheduledStates *int64, err error)
	addStateToHistory(stateName StateName, unscheduledStateName *StateName, numOfScheduledStates *int64) error
	removeLastStateFromHistory() error
	deleteStateHistory() error
	stateHistoryAvailable() bool
}

type defaultStateHistoryStash struct {
	JSONManager
}

// addStateToHistory adds a stateDetail to the history. Specify the values for unscheduledState and numOfScheduledStates to
// maintain the list of scheduled states if sub-flows are involved.
func (s *defaultStateHistoryStash) addStateToHistory(stateName StateName, unscheduledStateName *StateName, numOfScheduledStates *int64) error {
	// Create a new JSONManager to manage the history item
	historyItem := NewJSONManager()

	// Get the last state from history
	lastStateName, _, _, err := s.getLastStateFromHistory()
	if err != nil {
		return err
	}

	// If the last state is the same as the current state, do not add it again
	if lastStateName != nil && *lastStateName == stateName {
		return nil
	}

	// Set the stateDetail in the history item
	if err = historyItem.Set("s", stateName); err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	// If numOfScheduledStates is provided and greater than 0, set it in the history item
	if numOfScheduledStates != nil && *numOfScheduledStates > 0 {
		if err = historyItem.Set("n", *numOfScheduledStates); err != nil {
			return fmt.Errorf("failed to set num_of_scheduled_states: %w", err)
		}
	}

	// If unscheduledStateName is provided, set it in the history item
	if unscheduledStateName != nil {
		if err = historyItem.Set("u", *unscheduledStateName); err != nil {
			return fmt.Errorf("failed to set unscheduled_state: %w", err)
		}
	}

	// Add the new history item to the history
	if err = s.Set("_.state_history.-1", historyItem.Unmarshal()); err != nil {
		return fmt.Errorf("failed to update stashed history: %w", err)
	}

	return nil
}

// removeLastStateFromHistory removes the last stateDetail from history.
func (s *defaultStateHistoryStash) removeLastStateFromHistory() error {
	if err := s.Delete("_.state_history.-1"); err != nil {
		return err
	}

	return nil
}

// getLastStateFromHistory returns the last stateDetail, as well as the values for unscheduledState and numOfScheduledStates.
// These values indicate that further states have been added or removed from the list of scheduled states during the
// last stateDetail.
func (s *defaultStateHistoryStash) getLastStateFromHistory() (stateName, unscheduledStateName *StateName, numOfScheduledStates *int64, err error) {
	// Get the index of the last history item
	lastItemPosition := s.Get("_.state_history.#").Int() - 1

	// Retrieve the last history item
	lastHistoryItem := s.Get(JSONManagerPath(fmt.Sprintf("_.state_history.%d", lastItemPosition)))

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

	// Parse 's' field and assign it to the 'stateDetail' variable
	sn := StateName(lastHistoryItem.Get("s").String())
	stateName = &sn

	// Check if 'u' field exists in the last history item
	if lastHistoryItem.Get("u").Exists() {
		// Parse 'u' field and assign it to the 'unscheduledState' variable
		usn := StateName(lastHistoryItem.Get("u").String())
		unscheduledStateName = &usn
	}

	// Check if 'n' field exists in the last history item
	if lastHistoryItem.Get("n").Exists() {
		// Parse 'n' field and assign it to the 'numOfScheduledStates' variable
		n := lastHistoryItem.Get("n").Int()
		numOfScheduledStates = &n
	}

	// Return the parsed values
	return stateName, unscheduledStateName, numOfScheduledStates, nil
}

func (s *defaultStateHistoryStash) deleteStateHistory() error {
	return s.Delete("_.state_history")
}

func (s *defaultStateHistoryStash) stateHistoryAvailable() bool {
	return len(s.Get("_.state_history").Array()) > 0
}
