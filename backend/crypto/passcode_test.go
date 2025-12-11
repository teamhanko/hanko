package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasscodeGenerator_Generate(t *testing.T) {
	pg := NewNumericPasscodeGenerator()
	passcode, err := pg.Generate()

	assert.NoError(t, err)
	assert.NotEmpty(t, passcode)
	assert.Equal(t, 6, len(passcode))
}

func TestPasscodeGenerator_Generate_Different_Codes(t *testing.T) {
	pg := NewNumericPasscodeGenerator()

	passcode1, err := pg.Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, passcode1)

	passcode2, err := pg.Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, passcode2)

	assert.NotEqual(t, passcode1, passcode2)
}

func TestAlphanumericPasscodeGenerator_Generate(t *testing.T) {
	pg := NewAlphanumericPasscodeGenerator()
	passcode, err := pg.Generate()

	assert.NoError(t, err)
	assert.NotEmpty(t, passcode)
	assert.Equal(t, 6, len(passcode))
}

func TestAlphanumericPasscodeGenerator_Generate_Different_Codes(t *testing.T) {
	pg := NewAlphanumericPasscodeGenerator()

	passcode1, err := pg.Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, passcode1)

	passcode2, err := pg.Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, passcode2)

	assert.NotEqual(t, passcode1, passcode2)
}
