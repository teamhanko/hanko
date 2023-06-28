package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPasscodeGenerator_Generate(t *testing.T) {
	pg := NewPasscodeGenerator()
	passcode, err := pg.Generate()

	assert.NoError(t, err)
	assert.NotEmpty(t, passcode)
	assert.Equal(t, 6, len(passcode))
}

func TestPasscodeGenerator_Generate_Different_Codes(t *testing.T) {
	pg := NewPasscodeGenerator()

	passcode1, err := pg.Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, passcode1)

	passcode2, err := pg.Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, passcode2)

	assert.NotEqual(t, passcode1, passcode2)
}
