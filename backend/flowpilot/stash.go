package flowpilot

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
)

const (
	stashKeyState           = "state"
	stashKeyPreviousState   = "prev_state"
	stashKeyScheduledStates = "scheduled"
	stashKeyData            = "data"
	stashKeyHistory         = "hist"
	stashKeyRevertible      = "revertible"
	stashKeySticky          = "sticky"
)

type stash interface {
	pushState(bool) error
	pushErrorState(StateName) error
	revertState() error
	isRevertible() bool
	getStateName() StateName
	getPreviousStateName() StateName
	addScheduledStateNames(...StateName)
	getNextStateName() StateName
	useCompression(bool)

	jsonmanager.JSONManager
}

type defaultStash struct {
	jm                  jsonmanager.JSONManager
	data                jsonmanager.JSONManager
	scheduledStateNames []StateName
	compressionEnabled  bool
}

// newStashFromJSONManager creates a new instance of stash with a given JSONManager.
func newStashFromJSONManager(jm jsonmanager.JSONManager) stash {
	data, _ := jsonmanager.NewJSONManagerFromString(jm.Get(stashKeyData).String())
	return &defaultStash{
		jm:                  jm,
		data:                data,
		scheduledStateNames: make([]StateName, 0),
		compressionEnabled:  false,
	}
}

// newStash creates a new instance of Stash with empty JSON data.
func newStash(nextStates ...StateName) (stash, error) {
	jm := jsonmanager.NewJSONManager()

	if len(nextStates) == 0 {
		return nil, errors.New("can't create a new stash without a state name")
	}

	if err := jm.Set(stashKeyState, nextStates[0]); err != nil {
		return nil, err
	}

	if err := jm.Set(stashKeyScheduledStates, reverseStateNames(nextStates[1:])); err != nil {
		return nil, err
	}

	if err := jm.Set(stashKeyData, "{}"); err != nil {
		return nil, err
	}

	return newStashFromJSONManager(jm), nil
}

// newStashFromString creates a new instance of Stash with the given JSON data.
func newStashFromString(data string) (stash, error) {
	var err error

	if len(data) > 0 && !startsWithCurlyBrace(data) {
		if data, err = decodeData(data); err != nil {
			return nil, fmt.Errorf("faiiled to decode stash data: %w", err)
		}
	}

	jm, err := jsonmanager.NewJSONManagerFromString(data)
	return newStashFromJSONManager(jm), err
}

func reverseStateNames(slice []StateName) []StateName {
	reversed := make([]StateName, len(slice))
	for i, v := range slice {
		reversed[len(slice)-1-i] = v
	}
	return reversed
}

func startsWithCurlyBrace(s string) bool {
	// Check if the string is not empty
	if len(s) == 0 {
		return false
	}
	// Check if the first character is '{'
	return s[0] == '{'
}

func encodeData(jsonData string) (string, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(jsonData)); err != nil {
		return "", err
	}

	if err := gw.Close(); err != nil {
		return "", err
	}

	gzippedData := buf.Bytes()
	base64GzippedData := base64.StdEncoding.EncodeToString(gzippedData)
	return base64GzippedData, nil
}

func decodeData(base64GzippedData string) (string, error) {
	gzippedData, err := base64.StdEncoding.DecodeString(base64GzippedData)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(gzippedData)
	gr, err := gzip.NewReader(buf)
	if err != nil {
		return "", err
	}

	defer gr.Close()

	decompressedData, err := io.ReadAll(gr)
	if err != nil {
		return "", err
	}

	return string(decompressedData), nil
}

// Get retrieves the value at the specified path in the JSON data.
func (h *defaultStash) Get(path string) gjson.Result {
	return h.data.Get(path)
}

// Set updates the JSON data at the specified path with the provided value.
func (h *defaultStash) Set(path string, value interface{}) error {
	return h.data.Set(path, value)
}

// Delete removes a value from the JSON data at the specified path.
func (h *defaultStash) Delete(path string) error {
	return h.data.Delete(path)
}

// String returns the JSON data as a string.
func (h *defaultStash) String() string {
	if h.compressionEnabled {
		s, _ := encodeData(h.jm.String())
		return s
	}
	return h.jm.String()
}

// Unmarshal parses the JSON data and returns it as an interface{}.
func (h *defaultStash) Unmarshal() interface{} {
	return h.jm.Unmarshal()
}

func (h *defaultStash) pushState(revertible bool) error {
	return h.push(h.data.String(), revertible, h.getStateName() != h.getNextStateName())
}

func (h *defaultStash) pushErrorState(nextState StateName) error {
	return h.push(h.jm.Get(stashKeyData).String(), h.isRevertible(), false, nextState)
}

func (h *defaultStash) push(newData string, revertible, writeHistory bool, nextStates ...StateName) error {
	var err error

	data := h.jm.Get(stashKeyData)
	scheduledStates := h.jm.Get(stashKeyScheduledStates)
	scheduledStatesArr := scheduledStates.Array()
	stateStr := h.jm.Get(stashKeyState).String()
	prevStateStr := h.jm.Get(stashKeyPreviousState).String()

	scheduledStatesUpdated := make([]StateName, len(scheduledStatesArr))
	maxIndex := len(scheduledStatesUpdated) - 1
	for index := range scheduledStatesUpdated {
		scheduledStatesUpdated[maxIndex-index] = StateName(scheduledStatesArr[index].String())
	}

	scheduledStatesUpdated = append(nextStates, append(h.scheduledStateNames, scheduledStatesUpdated...)...)
	if len(scheduledStatesUpdated) == 0 {
		return errors.New("no state left to be used as the next state")
	}

	nextStateName := scheduledStatesUpdated[0]
	scheduledStatesUpdated = reverseStateNames(scheduledStatesUpdated[1:])

	if writeHistory {
		histItem := "{}"
		for key, value := range map[string]interface{}{
			stashKeyState:           stateStr,
			stashKeyPreviousState:   prevStateStr,
			stashKeyData:            data.Value(),
			stashKeyRevertible:      revertible,
			stashKeyScheduledStates: scheduledStates.Value(),
		} {
			if histItem, err = sjson.Set(histItem, key, value); err != nil {
				return err
			}
		}

		stashKeyNewHistItem := fmt.Sprintf("%s.-1", stashKeyHistory)
		if err = h.jm.Set(stashKeyNewHistItem, gjson.Parse(histItem).Value()); err != nil {
			return err
		}
	}

	for key, value := range map[string]interface{}{
		stashKeyState:           nextStateName,
		stashKeyPreviousState:   stateStr,
		stashKeyData:            gjson.Parse(newData).Value(),
		stashKeyScheduledStates: scheduledStatesUpdated,
	} {
		if err = h.jm.Set(key, value); err != nil {
			return err
		}
	}

	return nil
}

func (h *defaultStash) revertState() error {
	var err error

	lastHistItemIndex := h.jm.Get(fmt.Sprintf("%s.#", stashKeyHistory)).Int() - 1
	lastHistItem := h.jm.Get(fmt.Sprintf("%s.%d", stashKeyHistory, lastHistItemIndex))

	if !lastHistItem.Exists() {
		return errors.New("no state to revert to")
	}

	if !lastHistItem.Get(stashKeyRevertible).Bool() {
		return errors.New("state is not revertible")
	}

	dataUpdated := lastHistItem.Get(stashKeyData)
	h.data.Get(stashKeySticky).ForEach(func(key, value gjson.Result) bool {
		path := fmt.Sprintf("%s.%s", stashKeySticky, key.String())
		updated, _ := sjson.Set(dataUpdated.String(), path, value.Value())
		dataUpdated = gjson.Parse(updated)
		return true
	})

	if err = h.jm.Delete(fmt.Sprintf("%s.-1", stashKeyHistory)); err != nil {
		return err
	}

	for key, value := range map[string]interface{}{
		stashKeyScheduledStates: lastHistItem.Get(stashKeyScheduledStates).Value(),
		stashKeyState:           lastHistItem.Get(stashKeyState).Value(),
		stashKeyPreviousState:   lastHistItem.Get(stashKeyPreviousState).Value(),
		stashKeyData:            dataUpdated.Value(),
	} {
		if err = h.jm.Set(key, value); err != nil {
			return err
		}
	}

	return nil
}

func (h *defaultStash) getStateName() StateName {
	return StateName(h.jm.Get(stashKeyState).String())
}

func (h *defaultStash) getPreviousStateName() StateName {
	return StateName(h.jm.Get(stashKeyPreviousState).String())
}

func (h *defaultStash) addScheduledStateNames(names ...StateName) {
	h.scheduledStateNames = append(h.scheduledStateNames, names...)
}

func (h *defaultStash) getNextStateName() StateName {
	if len(h.scheduledStateNames) > 0 {
		return h.scheduledStateNames[0]
	}

	lastScheduledIndex := h.jm.Get(fmt.Sprintf("%s.#", stashKeyScheduledStates)).Int() - 1
	return StateName(h.jm.Get(fmt.Sprintf("%s.%d", stashKeyScheduledStates, lastScheduledIndex)).String())
}

func (h *defaultStash) isRevertible() bool {
	lastHistItemIndex := h.jm.Get(fmt.Sprintf("%s.#", stashKeyHistory)).Int() - 1
	return h.jm.Get(fmt.Sprintf("%s.%d.%s", stashKeyHistory, lastHistItemIndex, stashKeyRevertible)).Bool()
}

func (h *defaultStash) useCompression(b bool) {
	h.compressionEnabled = b
}
