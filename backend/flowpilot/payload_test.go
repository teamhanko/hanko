package flowpilot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/flowpilot/jsonmanager"
)

func Test_newPayload(t *testing.T) {
	// Create a new payload instance
	p := newPayload()

	// Assert that the payload is of the correct type
	assert.Implements(t, (*payload)(nil), p, "newPayload() should return an instance of payload")

	// Optionally, you can check if the returned payload is a JSONManager
	_, ok := p.(jsonmanager.JSONManager)
	if !ok {
		t.Errorf("newPayload() returned an instance that does not implement jsonmanager.JSONManager")
	}
}
